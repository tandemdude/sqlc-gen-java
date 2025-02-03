package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
	"github.com/tandemdude/sqlc-gen-java/internal/codegen"
	"github.com/tandemdude/sqlc-gen-java/internal/core"
	"github.com/tandemdude/sqlc-gen-java/internal/sql_types"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

// TODO - consider sqlc.embed support

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

	var typeConversionFunc sql_types.TypeConversionFunc
	switch req.Settings.Engine {
	case "postgresql":
		typeConversionFunc = sql_types.PostgresTypeToJavaType
	default:
		return nil, fmt.Errorf("engine %q is not supported", req.Settings.Engine)
	}

	// parse the incoming generate request into our Queries type
	var queries core.Queries = make(map[string][]core.Query)
	for _, query := range req.Queries {
		if _, ok := queries[query.Filename]; !ok {
			queries[query.Filename] = make([]core.Query, 0)
		}

		command, err := core.QueryCommandFor(query.Cmd)
		if err != nil {
			return nil, err
		}

		// TODO - check for array types? enum types? other specialness?
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
				Name:   arg.Column.Name,
				JavaType: core.JavaType{
					SqlType:  sdk.DataType(arg.Column.Type),
					Type:     javaType,
					IsList:   arg.Column.ArrayDims > 0, // TODO check this will always be present
					Nullable: !arg.Column.NotNull,
				},
			})
		}

		// TODO - check for array types? enum types? other specialness?
		returns := make([]core.QueryReturn, 0)
		for _, ret := range query.Columns {
			javaType, err := typeConversionFunc(ret.Type)
			if err != nil {
				return nil, err
			}

			if ret.ArrayDims > 1 {
				return nil, fmt.Errorf("multidimensional arrays are not supported, store JSON instead")
			}

			returns = append(returns, core.QueryReturn{
				Name: ret.Name,
				JavaType: core.JavaType{
					SqlType:  sdk.DataType(ret.Type),
					Type:     javaType,
					IsList:   ret.ArrayDims > 0, // TODO check this will always be present
					Nullable: !ret.NotNull,
				},
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
		fileName, fileContents, err := codegen.BuildQueriesFile(conf, file, queries[file])
		if err != nil {
			return nil, err
		}
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContents,
		})
	}

	// TODO - figure out common output models so we don't duplicate the same model in code 100 times

	return &plugin.GenerateResponse{Files: outputFiles}, nil
}
