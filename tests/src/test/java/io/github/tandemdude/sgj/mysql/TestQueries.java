package io.github.tandemdude.sgj.mysql;

import io.github.tandemdude.sgj.mysql.enums.BooksBookType;
import io.github.tandemdude.sgj.mysql.enums.NullableEnumTestEnumField;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.testcontainers.containers.MySQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.time.LocalDateTime;

import static org.assertj.core.api.Assertions.assertThat;

@Testcontainers
public class TestQueries {
    @Container
    private final MySQLContainer<?> mysql = new MySQLContainer<>("mysql:latest")
            .withInitScript("mysql/schema.sql");

    Connection getConn() throws SQLException {
        var conn = DriverManager.getConnection(mysql.getJdbcUrl(), mysql.getUsername(), mysql.getPassword());
        conn.setAutoCommit(true);
        return conn;
    }

    @Test
    @DisplayName("enum types can be read and written")
    void enumTypesCanBeReadAndWritten() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var authorId = q.createAuthor("foo");
            var bookId = q.createBook((int) authorId, "foo", BooksBookType.FICTION, "bar", 2000, LocalDateTime.now(), "baz");

            var foundRow = q.getBook((int) bookId);
            assertThat(foundRow).isPresent();
            var found = foundRow.get();
            assertThat(found.bookType()).isEqualTo(BooksBookType.FICTION);
        }
    }

    @Test
    @DisplayName("nullable enum types can be read and written")
    void nullableEnumTypesCanBeReadAndWritten() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var id = q.createEnumRow(NullableEnumTestEnumField.BAR);
            var foundRow = q.getEnumRow((int) id);
            assertThat(foundRow).isPresent();
            assertThat(foundRow.get().enumField()).isEqualTo(NullableEnumTestEnumField.BAR);

            var id2 = q.createEnumRow(null);
            var foundRow2 = q.getEnumRow((int) id2);
            assertThat(foundRow2).isPresent();
            assertThat(foundRow2.get().enumField()).isNull();
        }
    }
}
