package dev.danielslab.shorturl.utils

import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*

fun ApplicationCall.requireUserId(): Int {
    val principal = principal<JWTPrincipal>()
        ?: error("Missing JWT principal")

    return principal.payload.getClaim("userId").asInt()
        ?: error("Missing userId claim")
}
