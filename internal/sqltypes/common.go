package sqltypes

import (
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/tandemdude/sqlc-gen-java/poet"
)

type TypeConversionFunc func(*plugin.Identifier) (poet.TypeName, error)
