package dev.danielslab.shorturl

import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.testing.*
import kotlin.test.Test
import kotlin.test.assertEquals

class ApplicationTest {

    @Test
    fun testHealthRoute() = testApplication {
        application {
            module()
        }
        client.get("/api/v1/health").apply {
            assertEquals(HttpStatusCode.OK, status)
            assertEquals("""{"status":"ok"}""", bodyAsText())
        }
    }

}
