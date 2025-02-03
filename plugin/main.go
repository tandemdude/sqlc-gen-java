package main

import (
	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	java "github.com/tandemdude/sqlc-gen-java/internal"
)

func main() {
	codegen.Run(java.Generate)
}
