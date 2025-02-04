package io.github.tandemdude.sgj.postgres;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.util.List;
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

    @Test
    @DisplayName("ListUsers returns empty list when no records found")
    public void listUsersReturnsEmptyListNoRecordsFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            assertThat(q.listUsers()).isEmpty();
        }
    }

    @Test
    @DisplayName("ListUsers returns populated list when records found")
    public void listUsersReturnsPopulatedListRecordsFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            q.createUser(UUID.randomUUID(), "foo", "bar");
            q.createUser(UUID.randomUUID(), "baz", "bork");

            var found = q.listUsers();
            assertThat(found).isNotEmpty();
            assertThat(found.size()).isEqualTo(2);
        }
    }

    @Test
    @DisplayName("CreateMessage processes input list correctly")
    public void createMessageProcessesInputListCorrectly() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var created = q.createMessage(1, UUID.randomUUID(), "foo", List.of("bar", "baz", "bork"));
            assertThat(created).isPresent();

            var found = q.getMessage(created.get().message_id());
            assertThat(found).isPresent();
            assertThat(found.get().attachments()).containsExactly("bar", "baz", "bork");
        }
    }
}
