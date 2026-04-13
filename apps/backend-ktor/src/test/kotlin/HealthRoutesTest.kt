package dev.danielslab.shorturl.routes

import dev.danielslab.shorturl.plugins.configureSerialization
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.routing.*
import io.ktor.server.testing.*
import kotlin.test.Test
import kotlin.test.assertEquals

class HealthRoutesTest {

    @Test
    fun testHealthRoute() = testApplication {
        application {
            configureSerialization()
            routing {
                route("/api/v1/health") {
                    healthRoutes()
                }
            }
        }

        client.get("/api/v1/health").apply {
            assertEquals(HttpStatusCode.OK, status)
            assertEquals("""{"status":"ok"}""", bodyAsText())
        }
    }
}
