package io.github.tandemdude.sgj.postgres;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.nio.charset.StandardCharsets;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

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
    void getUserReturnsEmptyOptionalNoRecordsFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            assertThat(q.getUser(UUID.randomUUID())).isEmpty();
        }
    }

    @Test
    @DisplayName("GetUser returns populated optional when record found")
    void getUserReturnsPopulatedOptionalRecordFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var uid = UUID.randomUUID();
            q.createUser(uid, "foo", "bar");

            var found = q.getUser(uid);
            assertThat(found).isPresent();
            assertThat(found.get().userId()).isEqualTo(uid);
            assertThat(found.get().username()).isEqualTo("foo");
            assertThat(found.get().email()).isEqualTo("bar");
        }
    }

    @Test
    @DisplayName("ListUsers returns empty list when no records found")
    void listUsersReturnsEmptyListNoRecordsFound() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            assertThat(q.listUsers()).isEmpty();
        }
    }

    @Test
    @DisplayName("ListUsers returns populated list when records found")
    void listUsersReturnsPopulatedListRecordsFound() throws Exception {
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
    @DisplayName("GetUserDup throws error when multiple records returned")
    void getUserDupReturnsErrorWhenMultipleRecordsReturned() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            q.createUser(UUID.randomUUID(), "foo", "bar");
            q.createUser(UUID.randomUUID(), "baz", "bork");

            assertThatThrownBy(q::getUserDup)
                .isInstanceOf(SQLException.class)
                .hasMessageStartingWith("expected one row in result set");
        }
    }

    @Test
    @DisplayName("CreateMessage processes input list correctly")
    void createMessageProcessesInputListCorrectly() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var created = q.createMessage(1, UUID.randomUUID(), "foo", List.of("bar", "baz", "bork"));
            assertThat(created).isPresent();

            var found = q.getMessage(created.get());
            assertThat(found).isPresent();
            assertThat(found.get().attachments()).containsExactly("bar", "baz", "bork");
        }
    }

    @Test
    @DisplayName("GetMessage works when attachments is null")
    void getMessageWorksWhenAttachmentsIsNull() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var created = q.createMessage(1, UUID.randomUUID(), "foo", null);
            assertThat(created).isPresent();

            var found = q.getMessage(created.get());
            assertThat(found).isPresent();
            assertThat(found.get().attachments()).isNull();
        }
    }

    @Test
    @DisplayName("GetUserAndToken returns embedded objects")
    void getUserAndTokenReturnsEmbeddedObjects() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var userUid = UUID.randomUUID();
            q.createUser(userUid, "foo", "bar");
            q.createToken(userUid, "token", LocalDateTime.now());

            var userAndToken = q.getUserAndToken(userUid);
            assertThat(userAndToken).isPresent();
            assertThat(userAndToken.get().user().username()).isEqualTo("foo");
            assertThat(userAndToken.get().token().userId()).isEqualTo(userUid);
        }
    }

    @Test
    @DisplayName("GetBytes returns same data as during creation")
    void getBytesReturnsSameDataAsDuringCreation() throws Exception {
        try (var conn = getConn()) {
            var q = new Queries(conn);

            var s1 = "foobar".getBytes(StandardCharsets.UTF_8);
            var r1 = q.createBytes(s1, null);
            assertThat(r1).isPresent();

            var found1 = q.getBytes(r1.get());
            assertThat(found1).isPresent();
            assertThat(found1.get().contents()).isEqualTo(s1);
            assertThat(found1.get().hash()).isNull();

            var s2 = "bazbork".getBytes(StandardCharsets.UTF_8);
            var r2 = q.createBytes(s1, s2);
            assertThat(r2).isPresent();

            var found2 = q.getBytes(r2.get());
            assertThat(found2).isPresent();
            assertThat(found2.get().contents()).isEqualTo(s1);
            assertThat(found2.get().hash()).isEqualTo(s2);
        }
    }
}
