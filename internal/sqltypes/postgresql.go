package sqltypes

import (
	"fmt"

	"github.com/tandemdude/sqlc-gen-java/poet"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func PostgresTypeToJavaType(identifier *plugin.Identifier) (string, error) {
	colType := sdk.DataType(identifier)

	switch colType {
	case "serial", "pg_catalog.serial4", "integer", "int", "int4", "pg_catalog.int4":
		return "Integer", nil
	case "bigserial", "pg_catalog.serial8", "bigint", "pg_catalog.int8":
		return "Long", nil
	case "smallserial", "pg_catalog.serial2", "smallint", "pg_catalog.int2":
		return "Short", nil
	case "float", "double precision", "pg_catalog.float8":
		return "Double", nil
	case "real", "pg_catalog.float4":
		return "Float", nil
	case "pg_catalog.numeric":
		return "java.math.BigDecimal", nil
	case "bool", "pg_catalog.bool":
		return "Boolean", nil
	case "bytea", "blob", "pg_catalog.bytea":
		return "byte[]", nil
	case "date":
		return "java.time.LocalDate", nil
	case "pg_catalog.time", "pg_catalog.timetz":
		return "java.time.LocalTime", nil
	case "pg_catalog.timestamp", "timestamp":
		return "java.time.LocalDateTime", nil
	case "pg_catalog.timestamptz", "timestamptz":
		return "java.time.OffsetDateTime", nil
	case "text", "pg_catalog.varchar", "pg_catalog.bpchar", "string":
		return "String", nil
	case "uuid":
		return "java.util.UUID", nil
	// TODO - figure out if these can be supported properly
	case "jsonb", "inet":
		return "String", nil
	default:
		// void, any
		return "", fmt.Errorf("datatype '%s' not currently supported", colType)
	}
}

func ConvertPostgresType(identifier *plugin.Identifier) (poet.TypeName, error) {
	colType := sdk.DataType(identifier)

	switch colType {
	case "serial", "pg_catalog.serial4", "integer", "int", "int4", "pg_catalog.int4":
		return poet.IntBoxed, nil
	case "bigserial", "pg_catalog.serial8", "bigint", "pg_catalog.int8":
		return poet.LongBoxed, nil
	case "smallserial", "pg_catalog.serial2", "smallint", "pg_catalog.int2":
		return poet.ShortBoxed, nil
	case "float", "double precision", "pg_catalog.float8":
		return poet.DoubleBoxed, nil
	case "real", "pg_catalog.float4":
		return poet.FloatBoxed, nil
	case "pg_catalog.numeric":
		return poet.NewClassName("java.math", "BigDecimal"), nil
	case "bool", "pg_catalog.bool":
		return poet.BoolBoxed, nil
	case "bytea", "blob", "pg_catalog.bytea":
		return poet.Byte.Array(), nil
	case "date":
		return poet.NewClassName("java.time", "LocalDate"), nil
	case "pg_catalog.time", "pg_catalog.timetz":
		return poet.NewClassName("java.time", "LocalTime"), nil
	case "pg_catalog.timestamp", "timestamp":
		return poet.NewClassName("java.time", "LocalDateTime"), nil
	case "pg_catalog.timestamptz", "timestamptz":
		return poet.NewClassName("java.time", "OffsetDateTime"), nil
	case "text", "pg_catalog.varchar", "pg_catalog.bpchar", "string":
		return poet.String, nil
	case "uuid":
		return poet.NewClassName("java.util", "UUID"), nil
	// TODO - figure out if these can be supported properly
	case "jsonb", "inet":
		return poet.String, nil
	default:
		// void, any
		return poet.TypeName{}, fmt.Errorf("datatype '%s' not currently supported", colType)
	}
}
