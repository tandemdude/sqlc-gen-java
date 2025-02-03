package io.github.tandemdude.sgj.postgres;

import org.junit.Test;
import org.junit.jupiter.api.DisplayName;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.sql.DriverManager;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;

@Testcontainers
public class TestQueries {
    @Container
    private final PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:latest")
            .withInitScript("postgres/schema.sql");

    @Test
    @DisplayName("GetUser returns empty optional when no records found")
    public void getUserReturnsEmptyOptionalNoRecordsFound() throws Exception {
        try (var conn = DriverManager.getConnection(postgres.getJdbcUrl())) {
            var q = new Queries(conn);

            assertThat(q.getUser(UUID.randomUUID())).isEmpty();
        }
    }
}
