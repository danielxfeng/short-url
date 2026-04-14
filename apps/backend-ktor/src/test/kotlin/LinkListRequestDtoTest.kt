package dev.danielslab.shorturl

import dev.danielslab.shorturl.dto.LinkListRequestDto
import kotlin.test.Test
import kotlin.test.assertEquals

class LinkListRequestDtoTest {
    private val defaultPageSize = 20
    private val maxPageSize = 100

    @Test
    fun `uses default page size when limit is omitted`() {
        val request =
            LinkListRequestDto(
                cursor = null,
                limit = null,
            )

        val result =
            request.toRepoParam(
                userId = 1,
                defaultPageSize = defaultPageSize,
                maxPageSize = maxPageSize,
            )

        assertEquals(1, result.userId)
        assertEquals(null, result.cursor)
        assertEquals(20, result.limit)
    }

    @Test
    fun `keeps in-range limit`() {
        val request =
            LinkListRequestDto(
                cursor = "cursor-1",
                limit = 50,
            )

        val result =
            request.toRepoParam(
                userId = 1,
                defaultPageSize = defaultPageSize,
                maxPageSize = maxPageSize,
            )

        assertEquals(1, result.userId)
        assertEquals("cursor-1", result.cursor)
        assertEquals(50, result.limit)
    }

    @Test
    fun `raises too-small limit to default page size`() {
        val request =
            LinkListRequestDto(
                cursor = "cursor-1",
                limit = 5,
            )

        val result =
            request.toRepoParam(
                userId = 1,
                defaultPageSize = defaultPageSize,
                maxPageSize = maxPageSize,
            )

        assertEquals(20, result.limit)
    }

    @Test
    fun `caps too-large limit at max page size`() {
        val request =
            LinkListRequestDto(
                cursor = "cursor-1",
                limit = 200,
            )

        val result =
            request.toRepoParam(
                userId = 1,
                defaultPageSize = defaultPageSize,
                maxPageSize = maxPageSize,
            )

        assertEquals(100, result.limit)
    }
}
