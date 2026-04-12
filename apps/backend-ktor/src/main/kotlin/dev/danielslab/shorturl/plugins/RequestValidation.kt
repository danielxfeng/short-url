package dev.danielslab.shorturl.plugins

import io.ktor.server.application.*
import io.ktor.server.plugins.requestvalidation.*

fun Application.configureRequestValidation() {
    install(RequestValidation) {
        validate<String> { bodyText ->
            if (!bodyText.startsWith("Hello")) {
                ValidationResult.Invalid("Body text should start with 'Hello'")
            } else {
                ValidationResult.Valid
            }
        }
    }
}
