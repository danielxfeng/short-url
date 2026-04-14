package dev.danielslab.shorturl.routes

import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get

fun Route.linkRoutes() {
    get("") {
        call.respond(mapOf("status" to "ok"))
    }
}
