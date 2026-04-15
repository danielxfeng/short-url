package dev.danielslab.shorturl

import dev.danielslab.shorturl.dto.LinkCreateRequestDto
import dev.danielslab.shorturl.dto.LinkDeleteRequestDto
import dev.danielslab.shorturl.dto.LinkListRequestDto
import dev.danielslab.shorturl.plugins.configureRequestValidation
import dev.danielslab.shorturl.plugins.configureSerialization
import dev.danielslab.shorturl.plugins.validateLinkDeleteRequest
import dev.danielslab.shorturl.plugins.validateLinkListRequest
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.plugins.BadRequestException
import io.ktor.server.plugins.requestvalidation.RequestValidationException
import io.ktor.server.plugins.statuspages.exception
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.routing.post
import io.ktor.server.routing.routing
import io.ktor.server.testing.testApplication
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class LinkRequestValidationTest {
    @Test
    fun `accepts valid link create request`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page",
                          "note": "my note"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.OK, response.status)
        }

    @Test
    fun `accepts null note`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page",
                          "note": null
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.OK, response.status)
        }

    @Test
    fun `accepts omitted note`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.OK, response.status)
        }

    @Test
    fun `accepts omitted code`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "original_url": "https://example.com/page",
                          "note": "my note"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.OK, response.status)
        }

    @Test
    fun `rejects blank code on create`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "",
                          "original_url": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must not be blank", response.bodyAsText())
        }

    @Test
    fun `rejects too long code on create`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "${"a".repeat(256)}",
                          "original_url": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must be at most 255 characters", response.bodyAsText())
        }

    @Test
    fun `rejects invalid code characters on create`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello_123",
                          "original_url": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must contain only alphabets, numbers, and dash", response.bodyAsText())
        }

    @Test
    fun `rejects invalid code type on create`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": {},
                          "original_url": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("code", ignoreCase = true))
        }

    @Test
    fun `rejects blank note`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page",
                          "note": ""
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("note must not be blank", response.bodyAsText())
        }

    @Test
    fun `rejects too long note`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page",
                          "note": "${"a".repeat(256)}"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("note must be at most 255 characters", response.bodyAsText())
        }

    @Test
    fun `rejects invalid note type`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://example.com/page",
                          "note": {}
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("note", ignoreCase = true))
        }

    @Test
    fun `rejects non-http url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "ftp://example.com/file"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects localhost url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://localhost:8080/admin"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects private ip url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://192.168.1.10/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects ipv6 loopback url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://[::1]/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects local domain url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://printer.local/settings"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects internal domain url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://service.internal/dashboard"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects home domain url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://gateway.home/"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects lan domain url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://nas.lan/"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects url with user info`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "https://user:pass@example.com/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects unspecified ipv4 url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://0.1.2.3/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects carrier grade nat url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://100.64.1.1/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects benchmark network url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://198.18.0.1/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects multicast ipv4 url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://224.0.0.1/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects unspecified ipv6 url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://[::]/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects multicast ipv6 url`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": "http://[ff02::1]/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
        }

    @Test
    fun `rejects invalid originalUrl type`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links") {
                    contentType(ContentType.Application.Json)
                    setBody(
                        """
                        {
                          "code": "hello-world",
                          "original_url": {}
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("original_url", ignoreCase = true))
        }

    @Test
    fun `accepts valid delete request`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links/delete") {
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":"hello-world"}""")
                }

            assertEquals(HttpStatusCode.OK, response.status)
        }

    @Test
    fun `delete helper accepts valid code`() {
        assertEquals(null, validateLinkDeleteRequest(LinkDeleteRequestDto(code = "hello-world-123")))
    }

    @Test
    fun `delete helper rejects blank code`() {
        assertEquals("code must not be blank", validateLinkDeleteRequest(LinkDeleteRequestDto(code = "")))
    }

    @Test
    fun `delete helper rejects too long code`() {
        assertEquals(
            "code must be at most 255 characters",
            validateLinkDeleteRequest(LinkDeleteRequestDto(code = "a".repeat(256))),
        )
    }

    @Test
    fun `delete helper rejects invalid characters`() {
        assertEquals(
            "code must contain only alphabets, numbers, and dash",
            validateLinkDeleteRequest(LinkDeleteRequestDto(code = "hello_123")),
        )
    }

    @Test
    fun `list helper accepts null cursor`() {
        assertEquals(null, validateLinkListRequest(LinkListRequestDto(cursor = null, limit = null)))
    }

    @Test
    fun `list helper accepts valid cursor`() {
        assertEquals(null, validateLinkListRequest(LinkListRequestDto(cursor = "123", limit = null)))
    }

    @Test
    fun `list helper rejects blank cursor`() {
        assertEquals(
            "cursor must not be blank, and at most 255 characters",
            validateLinkListRequest(LinkListRequestDto(cursor = "", limit = null)),
        )
    }

    @Test
    fun `list helper rejects too long cursor`() {
        assertEquals(
            "cursor must not be blank, and at most 255 characters",
            validateLinkListRequest(LinkListRequestDto(cursor = "a".repeat(256), limit = null)),
        )
    }
}

private fun Application.linkValidationTestModule() {
    configureSerialization()
    configureRequestValidation()

    install(io.ktor.server.plugins.statuspages.StatusPages) {
        exception<RequestValidationException> { call, cause ->
            call.respond(HttpStatusCode.BadRequest, cause.reasons.joinToString())
        }

        exception<BadRequestException> { call, cause ->
            val message =
                generateSequence(cause as Throwable?) { it.cause }
                    .mapNotNull { it.message }
                    .joinToString(" | ")
                    .ifBlank { "Bad request" }

            call.respond(HttpStatusCode.BadRequest, message)
        }

        exception<Throwable> { call, _ ->
            call.respond(HttpStatusCode.InternalServerError)
        }
    }

    routing {
        post("/links") {
            call.receive<LinkCreateRequestDto>()
            call.respond(HttpStatusCode.OK)
        }

        post("/links/delete") {
            call.receive<LinkDeleteRequestDto>()
            call.respond(HttpStatusCode.OK)
        }
    }
}
