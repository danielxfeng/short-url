package dev.danielslab.shorturl.routes

import dev.danielslab.shorturl.plugins.configureSerialization
import io.ktor.client.request.get
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.server.routing.route
import io.ktor.server.routing.routing
import io.ktor.server.testing.testApplication
import kotlin.test.Test
import kotlin.test.assertEquals

class HealthRoutesTest {
    @Test
    fun testHealthRoute() =
        testApplication {
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
