package dev.danielslab.shorturl.plugins

import dev.danielslab.shorturl.dto.LinkCreateRequestDto
import dev.danielslab.shorturl.dto.LinkDeleteRequestDto
import dev.danielslab.shorturl.dto.UserUpsertRequestDto
import io.ktor.server.application.*
import io.ktor.server.plugins.requestvalidation.*
import java.net.URI

fun Application.configureRequestValidation() {
    install(RequestValidation) {
        validate<LinkCreateRequestDto> { request ->
            when {
                validateRequiredString(request.code, "code", maxLength = 255) != null ->
                    ValidationResult.Invalid(validateRequiredString(request.code, "code", maxLength = 255)!!)

                validateOptionalString(request.note, "note", maxLength = 255, combineBlankAndMax = false) != null ->
                    ValidationResult.Invalid(validateOptionalString(request.note, "note", maxLength = 255, combineBlankAndMax = false)!!)

                !isAllowedPublicHttpUrl(request.originalUrl) ->
                    ValidationResult.Invalid("originalUrl must be a public http/https URL")

                else -> ValidationResult.Valid
            }
        }

        validate<LinkDeleteRequestDto> { request ->
            when {
                validateRequiredString(request.code, "code", maxLength = 255) != null ->
                    ValidationResult.Invalid(validateRequiredString(request.code, "code", maxLength = 255)!!)

                else -> ValidationResult.Valid
            }
        }

        validate<UserUpsertRequestDto> { request ->
            when {
                validateRequiredString(request.providerId, "providerId") != null ->
                    ValidationResult.Invalid(validateRequiredString(request.providerId, "providerId")!!)

                validateOptionalString(request.displayName, "displayName", maxLength = 255) != null ->
                    ValidationResult.Invalid(validateOptionalString(request.displayName, "displayName", maxLength = 255)!!)

                validateOptionalString(request.profilePicture, "profilePicture", maxLength = 2048) != null ->
                    ValidationResult.Invalid(validateOptionalString(request.profilePicture, "profilePicture", maxLength = 2048)!!)

                else -> ValidationResult.Valid
            }
        }
    }
}

private fun validateRequiredString(
    value: String,
    fieldName: String,
    minLength: Int = 1,
    maxLength: Int? = null,
): String? {
    if (value.isBlank()) return "$fieldName must not be blank"
    if (value.length < minLength) return "$fieldName must be at least $minLength characters"
    if (maxLength != null && value.length > maxLength) {
        return "$fieldName must be at most $maxLength characters"
    }
    return null
}

private fun validateOptionalString(
    value: String?,
    fieldName: String,
    minLength: Int = 1,
    maxLength: Int? = null,
    combineBlankAndMax: Boolean = true,
): String? {
    if (value == null) return null
    if (value.isBlank()) {
        return if (maxLength == null || !combineBlankAndMax) {
            "$fieldName must not be blank"
        } else {
            "$fieldName must not be blank, and at most $maxLength characters"
        }
    }
    if (value.length < minLength) return "$fieldName must be at least $minLength characters"
    if (maxLength != null && value.length > maxLength) {
        return if (minLength <= 1 && combineBlankAndMax) {
            "$fieldName must not be blank, and at most $maxLength characters"
        } else if (minLength <= 1) {
            "$fieldName must be at most $maxLength characters"
        } else {
            "$fieldName must be between $minLength and $maxLength characters"
        }
    }
    return null
}

private fun isAllowedPublicHttpUrl(value: String): Boolean {
    val uri = runCatching { URI(value) }.getOrNull() ?: return false
    val scheme = uri.scheme?.lowercase() ?: return false
    if (scheme != "http" && scheme != "https") return false
    if (!uri.userInfo.isNullOrBlank()) return false

    val host = uri.host?.lowercase() ?: return false
    if (host == "localhost" || host.endsWith(".localhost") || host.endsWith(".local")) return false
    if (host.startsWith("[") && host.endsWith("]")) {
        return !isPrivateIpLiteral(host.removePrefix("[").removeSuffix("]"))
    }

    return !isPrivateIpLiteral(host)
}

private fun isPrivateIpLiteral(host: String): Boolean {
    if (":" in host) {
        val normalized = host.lowercase()
        return normalized == "::1" ||
            normalized.startsWith("fc") ||
            normalized.startsWith("fd") ||
            normalized.startsWith("fe80:")
    }

    val octets = host.split(".")
    if (octets.size != 4 || octets.any { it.isEmpty() || it.toIntOrNull() == null }) {
        return false
    }

    val values = octets.map { it.toInt() }
    if (values.any { it !in 0..255 }) return false

    return values[0] == 10 ||
        values[0] == 127 ||
        (values[0] == 169 && values[1] == 254) ||
        (values[0] == 172 && values[1] in 16..31) ||
        (values[0] == 192 && values[1] == 168)
}
