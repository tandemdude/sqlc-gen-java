package inflection

import (
	"github.com/jinzhu/inflection"
	"strings"
)

func Singular(s string) string {
	// Manual fix for incorrect handling of "campus"
	//
	// https://github.com/kyleconroy/sqlc/issues/430
	// https://github.com/jinzhu/inflection/issues/13
	if strings.ToLower(s) == "campus" {
		return s
	}
	// Manual fix for incorrect handling of "meta"
	//
	// https://github.com/kyleconroy/sqlc/issues/1217
	// https://github.com/jinzhu/inflection/issues/21
	if strings.ToLower(s) == "meta" {
		return s
	}
	return inflection.Singular(s)
}
