package dev.danielslab.shorturl

import dev.danielslab.shorturl.dto.UserUpsertRequestDto
import dev.danielslab.shorturl.plugins.configureRequestValidation
import dev.danielslab.shorturl.plugins.configureSerialization
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.*
import io.ktor.server.plugins.requestvalidation.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import io.ktor.server.testing.*
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class UserRequestValidationTest {
    // Coverage matrix from the prompt:
    // provider: happy / missing / invalid enum / invalid type
    // providerId: happy / missing / blank / invalid type
    // displayName: happy(null) / happy(non-null) / happy(omitted) / blank / too long / invalid type
    // profilePicture: happy(null) / happy(non-null) / happy(omitted) / blank / too long / invalid type

    @Test
    fun `accepts valid user payload with provider and providerId`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "displayName": "Daniel",
                  "profilePicture": "https://example.com/avatar.png"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `accepts null displayName and null profilePicture`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GITHUB",
                  "providerId": "github-123",
                  "displayName": null,
                  "profilePicture": null
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `accepts omitted displayName`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GITHUB",
                  "providerId": "github-123",
                  "profilePicture": "https://example.com/avatar.png"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `accepts omitted profilePicture`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GITHUB",
                  "providerId": "github-123",
                  "displayName": "Daniel"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `rejects missing provider`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "providerId": "github-123"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("provider", ignoreCase = true))
    }

    @Test
    fun `rejects invalid provider enum`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "FACEBOOK",
                  "providerId": "facebook-123"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("provider", ignoreCase = true))
    }

    @Test
    fun `rejects invalid provider type`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": 123,
                  "providerId": "google-123"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("provider", ignoreCase = true))
    }

    @Test
    fun `rejects missing providerId`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("providerId", ignoreCase = true))
    }

    @Test
    fun `rejects blank providerId`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": ""
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertEquals("providerId must not be blank", response.bodyAsText())
    }

    @Test
    fun `rejects invalid providerId type`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": {}
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("providerId", ignoreCase = true))
    }

    @Test
    fun `accepts non-null displayName`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "displayName": "Valid Name"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `rejects blank displayName`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "displayName": ""
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertEquals("displayName must not be blank, and at most 255 characters", response.bodyAsText())
    }

    @Test
    fun `rejects invalid displayName type`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "displayName": {}
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("displayName", ignoreCase = true))
    }

    @Test
    fun `rejects too long displayName`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "displayName": "${"a".repeat(256)}"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertEquals("displayName must not be blank, and at most 255 characters", response.bodyAsText())
    }

    @Test
    fun `accepts non-null profilePicture`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "profilePicture": "https://example.com/avatar.png"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.OK, response.status)
    }

    @Test
    fun `rejects blank profilePicture`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "profilePicture": ""
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertEquals("profilePicture must not be blank, and at most 2048 characters", response.bodyAsText())
    }

    @Test
    fun `rejects too long profilePicture`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "profilePicture": "${"a".repeat(2049)}"
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertEquals("profilePicture must not be blank, and at most 2048 characters", response.bodyAsText())
    }

    @Test
    fun `rejects invalid profilePicture type`() = testApplication {
        application { validationTestModule() }

        val response = client.post("/users") {
            contentType(ContentType.Application.Json)
            setBody(
                """
                {
                  "provider": "GOOGLE",
                  "providerId": "google-123",
                  "profilePicture": {}
                }
                """.trimIndent()
            )
        }

        assertEquals(HttpStatusCode.BadRequest, response.status)
        assertTrue(response.bodyAsText().contains("profilePicture", ignoreCase = true))
    }

}

private fun Application.validationTestModule() {
    configureSerialization()
    configureRequestValidation()

    install(io.ktor.server.plugins.statuspages.StatusPages) {
        exception<RequestValidationException> { call, cause ->
            call.respond(HttpStatusCode.BadRequest, cause.reasons.joinToString())
        }

        exception<BadRequestException> { call, cause ->
            val message = generateSequence(cause as Throwable?) { it.cause }
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
        post("/users") {
            call.receive<UserUpsertRequestDto>()
            call.respond(HttpStatusCode.OK)
        }
    }
}
