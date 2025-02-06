package core

type Config struct {
	Package                     string   `json:"package"`
	EmitExactTableNames         bool     `json:"emit_exact_table_names"`
	InflectionExcludeTableNames []string `json:"inflection_exclude_table_names"`
	// TODO - implement support for this
	QueryParameterLimit int    `json:"query_parameter_limit"`
	IndentChar          string `json:"indent_char"`
	CharsPerIndentLevel int    `json:"chars_per_indent_level"`
	NullableAnnotation  string `json:"nullable_annotation"`
	NonNullAnnotation   string `json:"non_null_annotation"`
}
