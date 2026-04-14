package dev.danielslab.shorturl.plugins

import dev.danielslab.shorturl.dto.ErrorResponse
import io.ktor.http.HttpStatusCode
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.plugins.ContentTransformationException
import io.ktor.server.plugins.requestvalidation.RequestValidationException
import io.ktor.server.plugins.statuspages.StatusPages
import io.ktor.server.plugins.statuspages.exception
import io.ktor.server.response.respond

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

        exception<ContentTransformationException> { call, cause ->
            call.respond(
                status = HttpStatusCode.BadRequest,
                message = ErrorResponse(cause.message ?: "Bad request"),
            )
        }

        exception<RequestValidationException> { call, cause ->
            call.respond(
                status = HttpStatusCode.BadRequest,
                message = ErrorResponse(cause.reasons.joinToString(", ")),
            )
        }

        exception<Throwable> { call, applicationError ->
            call.application.environment.log
                .error("Unhandled application error", applicationError)
            call.respond(
                status = HttpStatusCode.InternalServerError,
                message = ErrorResponse("Internal server error"),
            )
        }
    }
}
