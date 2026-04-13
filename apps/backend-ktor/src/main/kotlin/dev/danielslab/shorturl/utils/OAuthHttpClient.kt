package dev.danielslab.shorturl.utils

import io.ktor.client.*
import io.ktor.client.engine.apache5.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.serialization.kotlinx.json.*

object OAuthHttpClient {
    val client: HttpClient = HttpClient(Apache5) {
        install(ContentNegotiation) {
            json()
        }
    }
}
