package sqltypes

import (
	"fmt"
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
	// TODO - figure out if this can be supported properly
	case "jsonb":
		return "String", nil
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
	case "inet", "void", "any":
		return "", fmt.Errorf("datatype '%s' not currently supported", colType)
	default:
		// TODO - deal with enums somehow
		return "", fmt.Errorf("datatype '%s' not currently supported", colType)
	}
}
