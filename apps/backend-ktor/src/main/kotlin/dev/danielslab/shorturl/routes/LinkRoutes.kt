package dev.danielslab.shorturl.routes

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.Link
import dev.danielslab.shorturl.dto.LinkCreateRequestDto
import dev.danielslab.shorturl.dto.LinkDeleteRequestDto
import dev.danielslab.shorturl.dto.LinkListRequestDto
import dev.danielslab.shorturl.dto.LinkResponseDto
import dev.danielslab.shorturl.dto.LinksResponseDto
import dev.danielslab.shorturl.plugins.validateLinkDeleteRequest
import dev.danielslab.shorturl.plugins.validateLinkListRequest
import dev.danielslab.shorturl.repository.core.LinkCreateInput
import dev.danielslab.shorturl.repository.core.LinkDeleteInput
import dev.danielslab.shorturl.repository.core.LinkListInput
import dev.danielslab.shorturl.repository.core.LinkRepository
import dev.danielslab.shorturl.utils.requireUserId
import io.ktor.http.HttpStatusCode
import io.ktor.http.URLBuilder
import io.ktor.server.application.ApplicationCall
import io.ktor.server.auth.authenticate
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.response.respondRedirect
import io.ktor.server.routing.Route
import io.ktor.server.routing.delete
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import io.ktor.server.routing.put
import kotlinx.coroutines.launch
import java.security.SecureRandom
import java.util.Base64

fun Route.linkRoutes(
    config: Config,
    linkRepository: LinkRepository,
) {
    get("/{code}") {
        val code = call.parameters["code"]

        if (code == null) {
            call.redirectInvalidUrl(config.notFoundPage, "code-is-missing")
            return@get
        }

        validateLinkDeleteRequest(LinkDeleteRequestDto(code = code))?.let {
            call.redirectInvalidUrl(config.notFoundPage, code)
            return@get
        }

        val link =
            try {
                linkRepository.getLinkByCode(code)
            } catch (e: Exception) {
                call.application.environment.log
                    .error("failed to get a link by code", e)
                call.redirectInvalidUrl(config.notFoundPage, code)
                return@get
            }

        if (link == null) {
            call.redirectInvalidUrl(config.notFoundPage, code)
            return@get
        }

        // async task, fire-and-forget
        call.application.launch {
            try {
                linkRepository.setLinkClicked(code)
            } catch (e: Exception) {
                call.application.environment.log
                    .error("failed to set link clicked", e)
            }
        }

        call.respondRedirect(link.originalUrl)
    }

    authenticate("auth-jwt") {
        get("") {
            val request =
                LinkListRequestDto(
                    cursor = call.parameters["cursor"],
                    limit = call.parameters["limit"],
                ).also { dto ->
                    validateLinkListRequest(dto)?.let { throw IllegalArgumentException(it) }
                }.toInput(
                    userId = call.requireUserId(),
                    defaultPageSize = config.linkDefaultPageSize,
                    maxPageSize = config.linkMaxPageSize,
                )

            val newRequest = request.copy(limit = request.limit + 1)
            val links = linkRepository.getLinksByUserId(newRequest)
            val page = links.take(request.limit)

            call.respond(
                LinksResponseDto(
                    links = page.map { it.toResponse() },
                    hasMore = links.size > request.limit,
                    cursor = page.lastOrNull()?.id,
                ),
            )
        }

        post("") {
            val request =
                call
                    .receive<LinkCreateRequestDto>()
                    .optionalGenerateRandomCode(
                        linkRepository = linkRepository,
                        retryCount = config.linkCodeGenerateRetries,
                        codeLength = config.linkCodeLength,
                    ).toInput(call.requireUserId())

            val link = linkRepository.createLink(request)

            call.respond(HttpStatusCode.Created, link.toResponse())
        }

        put("/{code}/restore") {
            linkRepository.restoreLinkById(call.linkDeleteInput())

            call.respond(HttpStatusCode.NoContent)
        }

        delete("/{code}") {
            linkRepository.softDeleteLinkById(call.linkDeleteInput())

            call.respond(HttpStatusCode.NoContent)
        }

        delete("/{code}/permanent") {
            linkRepository.permanentlyDeleteLinkById(call.linkDeleteInput())

            call.respond(HttpStatusCode.NoContent)
        }
    }
}

private suspend fun ApplicationCall.redirectInvalidUrl(
    host: String,
    code: String,
) {
    respondRedirect(
        URLBuilder(host)
            .apply {
                parameters.append("invalid-url", code)
            }.build(),
    )
}

private fun ApplicationCall.linkDeleteInput(): LinkDeleteInput =
    LinkDeleteRequestDto(
        code = parameters["code"] ?: throw IllegalArgumentException("code is missing"),
    ).also { request ->
        validateLinkDeleteRequest(request)?.let { throw IllegalArgumentException(it) }
    }.toInput(requireUserId())

private suspend fun LinkCreateRequestDto.optionalGenerateRandomCode(
    linkRepository: LinkRepository,
    retryCount: Int,
    codeLength: Int,
): LinkCreateRequestDto {
    code?.let { return this }

    repeat(retryCount) {
        val generatedCode = generateRandomCode(codeLength)
        if (linkRepository.getLinkByCodeWithDeleted(generatedCode) == null) {
            return copy(code = generatedCode)
        }
    }

    throw RuntimeException("failed to generate an unique code")
}

// https://github.com/go-chi/chi/blob/master/middleware/request_id.go
private fun generateRandomCode(length: Int = 8): String {
    val random = SecureRandom()
    val replacer = Regex("[+/]")
    val buffer = ByteArray(12)
    var encoded = ""

    while (encoded.length < length) {
        random.nextBytes(buffer)
        encoded =
            Base64
                .getEncoder()
                .encodeToString(buffer)
                .replace(replacer, "")
    }

    return encoded.take(length)
}

private fun LinkCreateRequestDto.toInput(userId: Int): LinkCreateInput =
    LinkCreateInput(
        userId = userId,
        code = code!!,
        originalUrl = originalUrl,
        note = note,
    )

private fun LinkListRequestDto.toInput(
    userId: Int,
    defaultPageSize: Int,
    maxPageSize: Int,
): LinkListInput =
    LinkListInput(
        userId = userId,
        cursor = cursor,
        limit =
            limit
                ?.toIntOrNull()
                ?.coerceIn(defaultPageSize, maxPageSize)
                ?: defaultPageSize,
    )

private fun LinkDeleteRequestDto.toInput(userId: Int): LinkDeleteInput = LinkDeleteInput(userId = userId, code = code)

private fun Link.toResponse(): LinkResponseDto =
    LinkResponseDto(
        id = id,
        userId = userId,
        code = code,
        originalUrl = originalUrl,
        clicks = clicks,
        note = note,
        createdAt = createdAt.toString(),
        isDeleted = deletedAt != null,
    )
