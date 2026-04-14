package dev.danielslab.shorturl

import dev.danielslab.shorturl.dto.LinkCreateRequestDto
import dev.danielslab.shorturl.dto.LinkDeleteRequestDto
import dev.danielslab.shorturl.plugins.configureRequestValidation
import dev.danielslab.shorturl.plugins.configureSerialization
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
                          "originalUrl": "https://example.com/page",
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
                          "originalUrl": "https://example.com/page",
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
                          "originalUrl": "https://example.com/page"
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
                          "originalUrl": "https://example.com/page"
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
                          "originalUrl": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must be at most 255 characters", response.bodyAsText())
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
                          "originalUrl": "https://example.com/page",
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
                          "originalUrl": "https://example.com/page",
                          "note": "${"a".repeat(256)}"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("note must be at most 255 characters", response.bodyAsText())
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
                          "originalUrl": "ftp://example.com/file"
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
                          "originalUrl": "http://localhost:8080/admin"
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
                          "originalUrl": "http://192.168.1.10/private"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("originalUrl must be a public http/https URL", response.bodyAsText())
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
                          "originalUrl": "https://example.com/page"
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("code", ignoreCase = true))
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
                          "originalUrl": "https://example.com/page",
                          "note": {}
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("note", ignoreCase = true))
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
                          "originalUrl": {}
                        }
                        """.trimIndent(),
                    )
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertTrue(response.bodyAsText().contains("originalUrl", ignoreCase = true))
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
    fun `rejects blank code on delete`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links/delete") {
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":""}""")
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must not be blank", response.bodyAsText())
        }

    @Test
    fun `rejects too long code on delete`() =
        testApplication {
            application { linkValidationTestModule() }

            val response =
                client.post("/links/delete") {
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":"${"a".repeat(256)}"}""")
                }

            assertEquals(HttpStatusCode.BadRequest, response.status)
            assertEquals("code must be at most 255 characters", response.bodyAsText())
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
