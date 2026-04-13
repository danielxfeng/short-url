package dev.danielslab.shorturl

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.dto.LinkListRequestDto
import kotlin.test.Test
import kotlin.test.assertEquals

class LinkListRequestDtoTest {
    private val config = Config(
        port = 8080,
        cors = "http://localhost:5173",
        backendPublicUrl = "http://localhost:8080",
        dbUrl = "jdbc:postgresql://localhost/default",
        testDbUrl = "jdbc:postgresql://localhost/test",
        jwtSecret = "secret",
        jwtExpiry = 168,
        notFoundPage = "http://localhost:5173/not-found",
        googleClientId = "google-client-id",
        googleClientSecret = "google-client-secret",
        githubClientId = "github-client-id",
        githubClientSecret = "github-client-secret",
        frontendRedirectUrl = "http://localhost:5173/auth/callback",
        linkDefaultPageSize = 20,
        linkMaxPageSize = 100,
    )

    @Test
    fun `uses default page size when limit is omitted`() {
        val request = LinkListRequestDto(
            cursor = null,
            limit = null,
        )

        val result = request.toRepoParam(config, userId = 1)

        assertEquals(1, result.userId)
        assertEquals(null, result.cursor)
        assertEquals(20, result.limit)
    }

    @Test
    fun `keeps in-range limit`() {
        val request = LinkListRequestDto(
            cursor = "cursor-1",
            limit = 50,
        )

        val result = request.toRepoParam(config, userId = 1)

        assertEquals(1, result.userId)
        assertEquals("cursor-1", result.cursor)
        assertEquals(50, result.limit)
    }

    @Test
    fun `raises too-small limit to default page size`() {
        val request = LinkListRequestDto(
            cursor = "cursor-1",
            limit = 5,
        )

        val result = request.toRepoParam(config, userId = 1)

        assertEquals(20, result.limit)
    }

    @Test
    fun `caps too-large limit at max page size`() {
        val request = LinkListRequestDto(
            cursor = "cursor-1",
            limit = 200,
        )

        val result = request.toRepoParam(config, userId = 1)

        assertEquals(100, result.limit)
    }
}
