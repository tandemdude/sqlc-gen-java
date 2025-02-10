package sqltypes

import "github.com/sqlc-dev/plugin-sdk-go/plugin"

type TypeConversionFunc func(*plugin.Identifier) (string, error)
