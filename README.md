# SQLC Gen Java

A WASM plugin for SQLC allowing the generation of Java code. Currently in development - not production ready.

> [!NOTE]
> Only the `PostgreSQL` engine is supported currently. Support for `MySQL` is planned.
 
## Configuration Values

| Name                     | Type    | Required | Description                                                                                                                              |
|--------------------------|---------|----------|------------------------------------------------------------------------------------------------------------------------------------------|
| `package`                | string  | yes      | The name of the package where the generated files will be located                                                                        |
| `query_parameter_limit`  | integer | no       | not yet implemented                                                                                                                      |
| `indent_char`            | string  | no       | The character to use to indent the code. Defaults to space `" "`                                                                         |
| `chars_per_indent_level` | integer | no       | The number of characters per indent level. Defaults to `4`                                                                               |
| `nullable_annotation`    | string  | no       | The full import path for the nullable annotation to use. Defaults to `org.jspecify.annotations.Nullable`. Set to empty string to disable |
| `non_null_annotation`    | string  | no       | The full import path for the nonnull annotation to use. Defaults to `org.jspecify.annotations.NonNull`. Set to empty string to disable   |

## Usage

`sqlc.yaml`
```yaml
version: '2'
plugins:
- name: java
  wasm:
    url: TODO
    sha256: TODO
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

Building the plugin is very simple, just clone the repository and run the following command within the `plugin` directory:
```bash
GOOS=wasip1 GOARCH=wasm go build -o ../sqlc-gen-java.wasm
```

A file `sqlc-gen-java.wasm` will be created in the repository root - you can then move it to your sqlc-enabled project
and reference the plugin in your `sqlc.yaml` file using `file://sqlc-gen-java.wasm` as the plugin URL.

You should ensure that the `sha256` value in your `sqlc.yaml` is correct for this new plugin file.
