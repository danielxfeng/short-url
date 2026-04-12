package dev.danielslab.shorturl.routes

import io.ktor.server.application.*
import io.ktor.server.routing.*

fun Application.configureRouting() {
    routing {
        route("/api/v1") {
            healthRoutes()
        }
    }
}
