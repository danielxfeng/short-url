package dev.danielslab.shorturl

import dev.danielslab.shorturl.domain.Link
import dev.danielslab.shorturl.dto.LinkResponse
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlin.time.Instant

class LinkResponseTest {
    @Test
    fun `maps non-deleted link to response with iso createdAt`() {
        val createdAt = Instant.parse("2026-04-13T10:15:30Z")
        val link = Link(
            id = 1,
            userId = 1,
            code = "abc123",
            originalUrl = "https://example.com",
            clicks = 7,
            note = "note",
            createdAt = createdAt,
            deletedAt = null,
        )

        val response = LinkResponse.fromDomain(link)

        assertEquals(1, response.id)
        assertEquals(1, response.userId)
        assertEquals("abc123", response.code)
        assertEquals("https://example.com", response.originalUrl)
        assertEquals(7, response.clicks)
        assertEquals("note", response.note)
        assertEquals("2026-04-13T10:15:30Z", response.createdAt)
        assertFalse(response.isDeleted)
    }

    @Test
    fun `maps deleted link to response with isDeleted true`() {
        val createdAt = Instant.parse("2026-04-13T10:15:30Z")
        val deletedAt = Instant.parse("2026-04-14T10:15:30Z")
        val link = Link(
            id = 2,
            userId = 2,
            code = "xyz789",
            originalUrl = "https://example.org",
            clicks = 0,
            note = null,
            createdAt = createdAt,
            deletedAt = deletedAt,
        )

        val response = LinkResponse.fromDomain(link)

        assertEquals("2026-04-13T10:15:30Z", response.createdAt)
        assertTrue(response.isDeleted)
    }
}
