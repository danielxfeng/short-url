package dev.danielslab.shorturl.plugins

import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.compression.*
import io.ktor.server.plugins.cors.routing.*
import io.ktor.server.plugins.hsts.*

fun Application.configureHTTP(
    corsHost: String,
    corsScheme: String,
) {
    install(Compression)
    install(CORS) {
        allowMethod(HttpMethod.Options)
        allowMethod(HttpMethod.Get)
        allowMethod(HttpMethod.Post)
        allowMethod(HttpMethod.Put)
        allowMethod(HttpMethod.Delete)
        allowHeader(HttpHeaders.Authorization)
        allowHost(corsHost, schemes = listOf(corsScheme))
    }
    install(HSTS) {
        includeSubDomains = true
    }
}
