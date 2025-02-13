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
	"github.com/tandemdude/sqlc-gen-java/internal/sqltypes"
)

var (
	defaultIndentChar          = " "
	defaultCharsPerIndentLevel = 4
	postgresPlaceholderRegexp  = regexp.MustCompile(`\B\$\d+\b`)
)

type JavaGenerator struct {
	req  *plugin.GenerateRequest
	conf core.Config

	queries core.Queries
	models  core.EmbeddedModels

	enums     core.Enums
	usedEnums []string

	typeConversionFunc sqltypes.TypeConversionFunc
	nullableHelpers    core.NullableHelpers
}

func NewJavaGenerator(req *plugin.GenerateRequest) (*JavaGenerator, error) {
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

	var typeConversionFunc sqltypes.TypeConversionFunc
	switch req.Settings.Engine {
	case "postgresql":
		typeConversionFunc = sqltypes.PostgresTypeToJavaType
	case "mysql":
		typeConversionFunc = sqltypes.MysqlTypeToJavaType
	default:
		return nil, fmt.Errorf("engine %q is not supported", req.Settings.Engine)
	}

	return &JavaGenerator{
		req:                req,
		conf:               conf,
		queries:            make(core.Queries),
		models:             make(core.EmbeddedModels),
		enums:              make(core.Enums),
		usedEnums:          make([]string, 0),
		typeConversionFunc: typeConversionFunc,
		nullableHelpers:    core.NullableHelpers{},
	}, nil
}

func (gen *JavaGenerator) fixQueryPlaceholders(query string) (string, error) {
	if gen.req.Settings.Engine != "postgresql" {
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

func (gen *JavaGenerator) parseQueryReturn(col *plugin.Column) (*core.QueryReturn, error) {
	isEnum := false
	strJavaType, err := gen.typeConversionFunc(col.Type)
	if err != nil {
		schema := col.Table.Schema
		if schema == "" {
			schema = gen.req.Catalog.DefaultSchema
		}

		enumQualifiedName := fmt.Sprintf("%s.%s", schema, col.Type.Name)
		if _, ok := gen.enums[enumQualifiedName]; !ok {
			return nil, err
		}

		gen.usedEnums = append(gen.usedEnums, enumQualifiedName)
		strJavaType = gen.conf.Package + ".enums." + codegen.EnumClassName(enumQualifiedName, gen.req.Catalog.DefaultSchema)
		isEnum = true
	}

	if col.ArrayDims > 1 {
		return nil, fmt.Errorf("multidimensional arrays are not supported, store JSON instead")
	}

	javaType := core.JavaType{
		SqlType:    sdk.DataType(col.Type),
		Type:       strJavaType,
		IsList:     col.IsArray,
		IsNullable: !col.NotNull,
		IsEnum:     isEnum,
	}

	if javaType.IsNullable {
		if javaType.IsList {
			gen.nullableHelpers.List = true
		} else {
			switch strJavaType {
			case "Integer":
				gen.nullableHelpers.Int = true
			case "Long":
				gen.nullableHelpers.Long = true
			case "Float":
				gen.nullableHelpers.Float = true
			case "Double":
				gen.nullableHelpers.Double = true
			case "Boolean":
				gen.nullableHelpers.Boolean = true
			}
		}
	}

	return &core.QueryReturn{
		Name:     strcase.ToLowerCamel(col.Name),
		JavaType: javaType,
	}, nil
}

func (gen *JavaGenerator) Run() (*plugin.GenerateResponse, error) {
	// parse out the enums from the generate request
	for _, schema := range gen.req.Catalog.Schemas {
		for _, enum := range schema.Enums {
			gen.enums[fmt.Sprintf("%s.%s", schema.Name, enum.Name)] = core.Enum{
				Schema: schema.Name,
				Name:   enum.Name,
				Values: enum.Vals,
			}
		}
	}

	// parse the incoming generate request into our Queries type
	for _, query := range gen.req.Queries {
		if _, ok := gen.queries[query.Filename]; !ok {
			gen.queries[query.Filename] = make([]core.Query, 0)
		}

		command, err := core.QueryCommandFor(query.Cmd)
		if err != nil {
			return nil, err
		}

		// TODO - enum types? other specialness?
		args := make([]core.QueryArg, 0)
		for _, arg := range query.Params {
			isEnum := false
			javaType, err := gen.typeConversionFunc(arg.Column.Type)
			if err != nil {
				// check if this is an enum type
				schema := arg.Column.Table.Schema
				if schema == "" {
					schema = gen.req.Catalog.DefaultSchema
				}

				enumQualifiedName := fmt.Sprintf("%s.%s", schema, arg.Column.Type.Name)
				if _, ok := gen.enums[enumQualifiedName]; !ok {
					return nil, err
				}

				gen.usedEnums = append(gen.usedEnums, enumQualifiedName)
				javaType = gen.conf.Package + ".enums." + codegen.EnumClassName(enumQualifiedName, gen.req.Catalog.DefaultSchema)
				isEnum = true
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
					IsList:     arg.Column.IsArray,
					IsNullable: !arg.Column.NotNull,
					IsEnum:     isEnum,
				},
			})
		}

		// TODO - enum types? other specialness?
		var returns []core.QueryReturn
		for _, ret := range query.Columns {
			if ret.EmbedTable == nil {
				// normal types
				qr, err := gen.parseQueryReturn(ret)
				if err != nil {
					return nil, errors.Join(errors.New("failed to parse query return column"), err)
				}

				returns = append(returns, *qr)
				continue
			}

			// handle embedded types
			var table *plugin.Table

			// find the catalog entry for the embedded table
			schema := gen.req.Catalog.DefaultSchema
			if ret.EmbedTable.Schema != "" {
				schema = ret.EmbedTable.Schema
			}

			for _, s := range gen.req.Catalog.Schemas {
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
			if !gen.conf.EmitExactTableNames {
				modelName = strcase.ToCamel(inflection.Singular(table.Rel.Name, gen.conf.InflectionExcludeTableNames))
			}

			// check if we already have an entry for this model
			if _, ok := gen.models[modelName]; !ok {
				var modelParams []core.QueryReturn
				for _, c := range table.Columns {
					qr, err := gen.parseQueryReturn(c)
					if err != nil {
						return nil, errors.Join(errors.New("failed to parse query return column"), err)
					}

					modelParams = append(modelParams, *qr)
				}

				gen.models[modelName] = modelParams
			}

			returns = append(returns, core.QueryReturn{
				Name: strcase.ToLowerCamel(modelName),
				JavaType: core.JavaType{
					SqlType: "",
					// we don't need to specify package here - models file will be generated in the same location as the queries file
					Type:       gen.conf.Package + ".models." + modelName,
					IsList:     false, // TODO - check: this *should* be impossible
					IsNullable: false, // TODO - check: empty record should be output instead
				},
				EmbeddedModel: &modelName,
			})
		}

		// TODO - look into fixing ? operator for postgresql JSONB operations maybe
		newQueryText, err := gen.fixQueryPlaceholders(query.Text)
		if err != nil {
			return nil, err
		}

		gen.queries[query.Filename] = append(gen.queries[query.Filename], core.Query{
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
	for file := range gen.queries {
		// order the queries for each file alphabetically
		slices.SortFunc(gen.queries[file], func(a, b core.Query) int { return strings.Compare(a.MethodName, b.MethodName) })

		// build the queries file contents
		fileName, fileContents, err := codegen.BuildQueriesFile(gen.req.Settings.Engine, gen.conf, file, gen.queries[file], gen.models, gen.nullableHelpers)
		if err != nil {
			return nil, err
		}
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContents,
		})
	}

	for modelName, model := range gen.models {
		fileName, fileContents, err := codegen.BuildModelFile(gen.conf, modelName, model)
		if err != nil {
			return nil, err
		}
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContents,
		})
	}

	// remove duplicate enum entries
	slices.Sort(gen.usedEnums)
	slices.Compact(gen.usedEnums)
	for _, qualName := range gen.usedEnums {
		if qualName == "" {
			continue
		}

		enum := gen.enums[qualName]
		fileName, fileContents, err := codegen.BuildEnumFile(gen.req.Settings.Engine, gen.conf, qualName, enum, gen.req.Catalog.DefaultSchema)
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

// TODO - check if the context is actually important for anything
func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	jg, err := NewJavaGenerator(req)
	if err != nil {
		return nil, err
	}

	return jg.Run()
}
