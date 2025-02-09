package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
	"github.com/tandemdude/sqlc-gen-java/internal/codegen"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/internal/inflection"
	"github.com/tandemdude/sqlc-gen-java/internal/sql_types"
)

var (
	defaultIndentChar          = " "
	defaultCharsPerIndentLevel = 4
	postgresPlaceholderRegexp  = regexp.MustCompile(`\B\$\d+\b`)
)

func fixQueryPlaceholders(engine, query string) (string, error) {
	if engine != "postgresql" {
		return query, nil
	}

	var placeholders []string
	newQuery := postgresPlaceholderRegexp.ReplaceAllStringFunc(query, func(s string) string {
		placeholders = append(placeholders, s)
		return "?"
	})

	for _, placeholder := range placeholders {
		if _, err := strconv.Atoi(strings.TrimPrefix(placeholder, "$")); err != nil {
			return "", fmt.Errorf("invalid placeholder in query: %s", placeholder)
		}
	}

	return newQuery, nil
}

func parseQueryReturn(tcf sql_types.TypeConversionFunc, nullableHelpers *core.NullableHelpers, col *plugin.Column) (*core.QueryReturn, error) {
	name := strcase.ToCamel(col.Name)
	strJavaType, err := tcf(col.Type)
	if err != nil {
		return nil, err
	}

	if col.ArrayDims > 1 {
		return nil, fmt.Errorf("multidimensional arrays are not supported, store JSON instead")
	}

	javaType := core.JavaType{
		SqlType:    sdk.DataType(col.Type),
		Type:       strJavaType,
		IsList:     col.IsArray,
		IsNullable: !col.NotNull,
	}

	if javaType.IsNullable {
		if javaType.IsList {
			nullableHelpers.List = true
		} else {
			switch strJavaType {
			case "Integer":
				nullableHelpers.Int = true
			case "Long":
				nullableHelpers.Long = true
			case "Float":
				nullableHelpers.Float = true
			case "Double":
				nullableHelpers.Double = true
			case "Boolean":
				nullableHelpers.Boolean = true
			}
		}
	}

	return &core.QueryReturn{
		Name:     strcase.ToLowerCamel(name),
		JavaType: javaType,
	}, nil
}

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	conf := core.Config{
		IndentChar:          defaultIndentChar,
		CharsPerIndentLevel: defaultCharsPerIndentLevel,
		NullableAnnotation:  "org.jspecify.annotations.Nullable",
		NonNullAnnotation:   "org.jspecify.annotations.NonNull",
	}
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return nil, err
		}
	}

	if conf.Package == "" {
		return nil, fmt.Errorf("'package' is a required configuration option")
	}

	var typeConversionFunc sql_types.TypeConversionFunc
	switch req.Settings.Engine {
	case "postgresql":
		typeConversionFunc = sql_types.PostgresTypeToJavaType
	default:
		return nil, fmt.Errorf("engine %q is not supported", req.Settings.Engine)
	}

	var queries core.Queries = make(map[string][]core.Query)
	var embeddedModels core.EmbeddedModels = make(map[string][]core.QueryReturn)
	var nullableHelpers core.NullableHelpers = core.NullableHelpers{}

	// parse the incoming generate request into our Queries type
	for _, query := range req.Queries {
		if _, ok := queries[query.Filename]; !ok {
			queries[query.Filename] = make([]core.Query, 0)
		}

		command, err := core.QueryCommandFor(query.Cmd)
		if err != nil {
			return nil, err
		}

		// TODO - enum types? other specialness?
		args := make([]core.QueryArg, 0)
		for _, arg := range query.Params {
			javaType, err := typeConversionFunc(arg.Column.Type)
			if err != nil {
				return nil, err
			}

			if arg.Column.ArrayDims > 1 {
				return nil, fmt.Errorf("multidimensional arrays are not supported, store JSON instead")
			}

			args = append(args, core.QueryArg{
				Number: int(arg.Number),
				Name:   strcase.ToLowerCamel(arg.Column.Name),
				JavaType: core.JavaType{
					SqlType:    sdk.DataType(arg.Column.Type),
					Type:       javaType,
					IsList:     arg.Column.IsArray, // TODO check this will always be present
					IsNullable: !arg.Column.NotNull,
				},
			})
		}

		// TODO - enum types? other specialness?
		var returns []core.QueryReturn
		for _, ret := range query.Columns {
			if ret.EmbedTable == nil {
				// normal types
				qr, err := parseQueryReturn(typeConversionFunc, &nullableHelpers, ret)
				if err != nil {
					return nil, errors.Join(errors.New("failed to parse query return column"), err)
				}

				returns = append(returns, *qr)
				continue
			}

			// handle embedded types
			var table *plugin.Table

			// find the catalog entry for the embedded table
			schema := req.Catalog.DefaultSchema
			if ret.EmbedTable.Schema != "" {
				schema = ret.EmbedTable.Schema
			}

			for _, s := range req.Catalog.Schemas {
				if s.Name != schema {
					continue
				}

				for _, t := range s.Tables {
					if ret.EmbedTable.Name == t.Rel.Name {
						table = t
						break
					}
				}
			}
			if table == nil {
				return nil, fmt.Errorf("unknown embedded table %s.%s", schema, ret.EmbedTable)
			}

			// TODO - fix type-writer to only exclude items that aren't part of the package name
			modelName := strcase.ToCamel(table.Rel.Name)
			if !conf.EmitExactTableNames {
				modelName = strcase.ToCamel(inflection.Singular(table.Rel.Name, conf.InflectionExcludeTableNames))
			}

			// check if we already have an entry for this model
			if _, ok := embeddedModels[modelName]; !ok {
				var modelParams []core.QueryReturn
				for _, c := range table.Columns {
					qr, err := parseQueryReturn(typeConversionFunc, &nullableHelpers, c)
					if err != nil {
						return nil, errors.Join(errors.New("failed to parse query return column"), err)
					}

					modelParams = append(modelParams, *qr)
				}

				embeddedModels[modelName] = modelParams
			}

			returns = append(returns, core.QueryReturn{
				Name: strcase.ToLowerCamel(modelName),
				JavaType: core.JavaType{
					SqlType: "",
					// we don't need to specify package here - models file will be generated in the same location as the queries file
					Type:       conf.Package + ".models." + modelName,
					IsList:     false, // TODO - check: this *should* be impossible
					IsNullable: false, // TODO - check: empty record should be output instead
				},
				EmbeddedModel: &modelName,
			})
		}

		// TODO - look into fixing ? operator for postgresql JSONB operations maybe
		newQueryText, err := fixQueryPlaceholders(req.Settings.Engine, query.Text)
		if err != nil {
			return nil, err
		}

		queries[query.Filename] = append(queries[query.Filename], core.Query{
			RawCommand:   query.Cmd,
			Command:      command,
			Text:         newQueryText,
			RawQueryName: query.Name,
			// TODO - clean the name of any disallowed characters?
			MethodName: strcase.ToLowerCamel(query.Name),
			Args:       args,
			Returns:    returns,
		})
	}

	outputFiles := make([]*plugin.File, 0)
	// order the queries for each file alphabetically
	for file := range queries {
		slices.SortFunc(queries[file], func(a, b core.Query) int { return strings.Compare(a.MethodName, b.MethodName) })

		// build the queries file contents
		fileName, fileContents, err := codegen.BuildQueriesFile(conf, file, queries[file], embeddedModels, nullableHelpers)
		if err != nil {
			return nil, err
		}
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContents,
		})
	}

	for modelName, model := range embeddedModels {
		fileName, fileContents, err := codegen.BuildModelFile(conf, modelName, model)
		if err != nil {
			return nil, err
		}
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContents,
		})
	}

	return &plugin.GenerateResponse{Files: outputFiles}, nil
}
