package dev.danielslab.shorturl.utils

import io.ktor.server.application.ApplicationCall
import io.ktor.server.auth.jwt.JWTPrincipal
import io.ktor.server.auth.principal

fun ApplicationCall.requireUserId(): Int {
    val principal =
        principal<JWTPrincipal>()
            ?: error("Missing JWT principal")

    return principal.payload.getClaim("userId").asInt()
        ?: error("Missing userId claim")
}
