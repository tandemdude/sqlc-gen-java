package io.github.tandemdude.sgj.postgres;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;

@Testcontainers
public class TestQueries {
    @Container
    private final PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:latest")
        .withInitScript("postgres/schema.sql");

    Connection getConn() throws SQLException {
        var conn = DriverManager.getConnection(postgres.getJdbcUrl(), postgres.getUsername(), postgres.getPassword());
        conn.setAutoCommit(true);
        return conn;
    }

    @Test
    @DisplayName("GetUser returns empty optional when no records found")
    public void getUserReturnsEmptyOptionalNoRecordsFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            assertThat(q.getUser(UUID.randomUUID())).isEmpty();
        }
    }

    @Test
    @DisplayName("GetUser returns populated optional when record found")
    public void getUserReturnsPopulatedOptionalRecordFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var uid = UUID.randomUUID();
            q.createUser(uid, "foo", "bar");

            var found = q.getUser(uid);
            assertThat(found).isPresent();
            assertThat(found.get().user_id()).isEqualTo(uid);
            assertThat(found.get().username()).isEqualTo("foo");
            assertThat(found.get().email()).isEqualTo("bar");
        }
    }
}
