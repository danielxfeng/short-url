package dev.danielslab.shorturl.plugins

import dev.danielslab.shorturl.dto.UserUpsertRequestDto
import io.ktor.server.application.*
import io.ktor.server.plugins.requestvalidation.*

fun Application.configureRequestValidation() {
    install(RequestValidation) {
        validate<UserUpsertRequestDto> { request ->
            when {
                request.providerId.isBlank() ->
                    ValidationResult.Invalid("providerId must not be blank")

                request.displayName != null && (request.displayName.isBlank() || request.displayName.length > 255) ->
                    ValidationResult.Invalid("displayName must not be blank, and at most 255 characters")

                request.profilePicture != null && (request.profilePicture.isBlank() || request.profilePicture.length > 2048) ->
                    ValidationResult.Invalid("profilePicture must not be blank, and at most 2048 characters")

                else -> ValidationResult.Valid
            }
        }
    }
}
