package sqltypes

import (
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func MysqlTypeToJavaType(identifier *plugin.Identifier) (string, error) {
	colType := sdk.DataType(identifier)

	switch colType {
	case "varchar", "text", "char", "tinytext", "mediumtext", "longtext":
		return "String", nil
	case "int", "integer", "smallint", "mediumint", "year":
		return "Integer", nil
	case "bigint":
		return "Long", nil
	case "blob", "binary", "varbinary", "tinyblob", "mediumblob", "longblob":
		return "byte[]", nil
	case "double", "double precision", "real":
		return "Double", nil
	case "decimal", "dec", "fixed":
		return "java.math.BigDecimal", nil
	case "date":
		return "java.time.LocalDate", nil
	case "datetime", "time":
		return "java.time.LocalDateTime", nil
	// TODO - instant support - look into option for this in pgsql as well
	case "timestamp":
		return "java.time.OffsetDateTime", nil
	case "boolean", "bool", "tinyint":
		return "Boolean", nil
	case "json":
		return "String", nil
	default:
		return "", fmt.Errorf("datatype '%s' not currently supported", colType)
	}
}
