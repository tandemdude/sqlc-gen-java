# SQLC Gen Java

A WASM plugin for SQLC allowing the generation of Java code.

> [!NOTE]
> Only the `PostgreSQL` engine is supported currently. Support for `MySQL` is planned.

> [!IMPORTANT]
> The generated code makes heavy use of records, so you must be using a Java version that has record support (14+). Support
> is not currently planned for earlier Java versions, but if you think you could implement it then feel free
> to open a pull request.

## Configuration Values

| Name                             | Type     | Required | Description                                                                                                                               |
|----------------------------------|----------|----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `package`                        | string   | yes      | The name of the package where the generated files will be located.                                                                        |
| `emit_exact_table_names`         | boolean  | no       | Whether table names will not be forced to singular form when generating the models. Defaults to `false`.                                  |
| `inflection_exclude_table_names` | []string | no       | Table names to be excluded from being forced into singular form when generating the models.                                               |
| `query_parameter_limit`          | integer  | no       | not yet implemented                                                                                                                       |
| `indent_char`                    | string   | no       | The character to use to indent the code. Defaults to space `" "`.                                                                         |
| `chars_per_indent_level`         | integer  | no       | The number of characters per indent level. Defaults to `4`.                                                                               |
| `nullable_annotation`            | string   | no       | The full import path for the nullable annotation to use. Defaults to `org.jspecify.annotations.Nullable`. Set to empty string to disable. |
| `non_null_annotation`            | string   | no       | The full import path for the nonnull annotation to use. Defaults to `org.jspecify.annotations.NonNull`. Set to empty string to disable.   |
| `expose_connection`              | boolean  | no       | Whether a getter will be generated for the internally held connection instance. Defaults to `false`.                                      |

## Usage

Check the [latest GitHub release](https://github.com/tandemdude/sqlc-gen-java/releases/latest) for the plugin download URL and checksum.

`sqlc.yaml`
```yaml
version: "2"
plugins:
  - name: java
    wasm:
      url: https://github.com/tandemdude/sqlc-gen-java/releases/download/{{VERSION}}/sqlc-gen-java.wasm
      sha256: {{CHECKSUM}}
sql:
  - schema: src/main/resources/postgresql/schema.sql
    queries: src/main/resources/postgresql/queries.sql
    engine: postgresql
    codegen:
      - out: src/main/java/com/example/postgresql
        plugin: java
        options:
          package: com.example.postgresql
```

## Building From Source

Building the plugin is very simple, just clone the repository and run the following command:

```bash
GOOS=wasip1 GOARCH=wasm go build -o sqlc-gen-java.wasm plugin/main.go
```

A file `sqlc-gen-java.wasm` will be created in the repository root - you can then move it to your sqlc-enabled project
and reference the plugin in your `sqlc.yaml` file using `file://sqlc-gen-java.wasm` as the plugin URL.

You should ensure that the `sha256` value in your `sqlc.yaml` is correct for this new plugin file.

## Planned Features

- `SQLite` support
- Improved parameter naming

**Tentative:**

- r2dbc support
- copyfrom support where possible [ref](https://www.baeldung.com/jdbc-batch-processing)
