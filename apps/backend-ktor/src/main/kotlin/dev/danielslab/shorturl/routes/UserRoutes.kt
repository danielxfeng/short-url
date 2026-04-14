package dev.danielslab.shorturl.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.dto.UserResponseDto
import dev.danielslab.shorturl.dto.oauth.GithubUserInfo
import dev.danielslab.shorturl.dto.oauth.GoogleUserInfo
import dev.danielslab.shorturl.dto.oauth.OAuthUserProfile
import dev.danielslab.shorturl.repository.core.NotFoundException
import dev.danielslab.shorturl.repository.core.UserRepository
import dev.danielslab.shorturl.utils.OAuthHttpClient
import dev.danielslab.shorturl.utils.requireUserId
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.URLBuilder
import io.ktor.server.application.ApplicationCall
import io.ktor.server.auth.OAuthAccessTokenResponse
import io.ktor.server.auth.authenticate
import io.ktor.server.auth.principal
import io.ktor.server.response.respond
import io.ktor.server.response.respondRedirect
import io.ktor.server.routing.Route
import io.ktor.server.routing.delete
import io.ktor.server.routing.get
import io.ktor.server.routing.route
import java.util.Date

fun Route.userRoutes(
    config: Config,
    userRepository: UserRepository,
) {
    route("/auth") {
        get("/{provider}") {
            val loginPath =
                when (call.parameters["provider"]?.lowercase()) {
                    "google" -> "/api/v1/user/auth/google/login"
                    "github" -> "/api/v1/user/auth/github/login"
                    else -> {
                        redirectAuthError(call, config.frontendRedirectUrl, "provider not found")
                        return@get
                    }
                }

            call.respondRedirect(loginPath)
        }

        authenticate("auth-oauth-google") {
            route("/google") {
                get("/login") {}

                get("/callback") {
                    val principal =
                        call.principal<OAuthAccessTokenResponse.OAuth2>()
                            ?: run {
                                redirectAuthError(call, config.frontendRedirectUrl, "failed to get token from google")
                                return@get
                            }

                    val user =
                        call.fetchAndUpsertOAuthUser<GoogleUserInfo>(
                            config = config,
                            userRepository = userRepository,
                            accessToken = principal.accessToken,
                            userInfoUrl = config.googleUserInfoUrl,
                        ) ?: return@get

                    val token = call.issueTokenOrRedirect(user.id, config) ?: return@get

                    call.respondRedirect(
                        URLBuilder(config.frontendRedirectUrl)
                            .apply {
                                parameters.append("auth", token)
                            }.build(),
                    )
                }
            }
        }

        authenticate("auth-oauth-github") {
            route("/github") {
                get("/login") {}

                get("/callback") {
                    val principal =
                        call.principal<OAuthAccessTokenResponse.OAuth2>()
                            ?: run {
                                redirectAuthError(call, config.frontendRedirectUrl, "failed to get token from github")
                                return@get
                            }

                    val user =
                        call.fetchAndUpsertOAuthUser<GithubUserInfo>(
                            config = config,
                            userRepository = userRepository,
                            accessToken = principal.accessToken,
                            userInfoUrl = config.githubUserInfoUrl,
                        ) ?: return@get

                    val token = call.issueTokenOrRedirect(user.id, config) ?: return@get

                    call.respondRedirect(
                        URLBuilder(config.frontendRedirectUrl)
                            .apply {
                                parameters.append("auth", token)
                            }.build(),
                    )
                }
            }
        }
    }

    authenticate("auth-jwt") {
        route("/me") {
            get("") {
                call.respond(
                    userRepository
                        .getUserById(call.requireUserId())
                        ?.toResponseDto()
                        ?: throw NotFoundException("User not found"),
                )
            }

            delete("") {
                val userId = call.requireUserId()
                userRepository.deleteUserById(userId)
                call.respond(HttpStatusCode.NoContent)
            }
        }
    }
}

suspend fun redirectAuthError(
    call: ApplicationCall,
    host: String,
    msg: String,
    cause: Throwable? = null,
) {
    if (cause != null) {
        call.application.environment.log
            .warn(msg, cause)
    } else {
        call.application.environment.log
            .warn(msg)
    }
    call.respondRedirect(
        URLBuilder(host)
            .apply {
                parameters.append("error", msg)
            }.build(),
    )
}

suspend fun ApplicationCall.issueTokenOrRedirect(
    userId: Int,
    config: Config,
): String? =
    try {
        JWT
            .create()
            .withAudience(config.jwtAudience)
            .withIssuer(config.jwtDomain)
            .withClaim("userId", userId)
            .withExpiresAt(Date(System.currentTimeMillis() + config.jwtExpiry * 60L * 60L * 1000L))
            .sign(Algorithm.HMAC256(config.jwtSecret))
    } catch (e: Exception) {
        redirectAuthError(this, config.frontendRedirectUrl, "failed to issue token", e)
        null
    }

suspend inline fun <reified T : OAuthUserProfile> ApplicationCall.fetchAndUpsertOAuthUser(
    config: Config,
    userRepository: UserRepository,
    accessToken: String,
    userInfoUrl: String,
): User? {
    val userInfo =
        try {
            OAuthHttpClient.client
                .get(userInfoUrl) {
                    header(HttpHeaders.Authorization, "Bearer $accessToken")
                }.body<T>()
        } catch (e: Exception) {
            redirectAuthError(this, config.frontendRedirectUrl, "failed to fetch user info", e)
            return null
        }

    return try {
        userRepository.upsertUser(userInfo.toUserUpsertInput())
    } catch (e: Exception) {
        redirectAuthError(this, config.frontendRedirectUrl, "failed to authorize", e)
        null
    }
}

private fun User.toResponseDto(): UserResponseDto =
    UserResponseDto(
        id = id,
        provider = provider,
        providerId = providerId,
        displayName = displayName,
        profilePicture = profilePicture,
    )
