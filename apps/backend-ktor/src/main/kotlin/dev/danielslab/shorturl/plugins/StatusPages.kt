package dev.danielslab.shorturl.plugins

import dev.danielslab.shorturl.dto.ErrorResponse
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.statuspages.*
import io.ktor.server.response.*

fun Application.configureStatusPages() {
    install(StatusPages) {
        exception<IllegalArgumentException> { call, cause ->
            call.respond(
                status = HttpStatusCode.BadRequest,
                message = ErrorResponse(cause.message ?: "Bad request"),
            )
        }

        exception<NoSuchElementException> { call, cause ->
            call.respond(
                status = HttpStatusCode.NotFound,
                message = ErrorResponse(cause.message ?: "Resource not found"),
            )
        }

        exception<IllegalStateException> { call, cause ->
            call.respond(
                status = HttpStatusCode.Conflict,
                message = ErrorResponse(cause.message ?: "Conflict"),
            )
        }

        exception<Throwable> { call, applicationError ->
            call.application.environment.log.error("Unhandled application error", applicationError)
            call.respond(
                status = HttpStatusCode.InternalServerError,
                message = ErrorResponse("Internal server error"),
            )
        }
    }
}
