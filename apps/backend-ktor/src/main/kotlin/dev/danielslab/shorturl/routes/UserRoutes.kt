package dev.danielslab.shorturl.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.dto.UserResponseDto
import dev.danielslab.shorturl.repository.core.NotFoundException
import dev.danielslab.shorturl.repository.core.UserRepository
import dev.danielslab.shorturl.repository.core.UserUpsertInput
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
import kotlinx.serialization.Serializable
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
                        redirectAuthError(call, config.notFoundPage, "provider not found")
                        return@get
                    }
                }

            call.respondRedirect(loginPath)
        }

        authenticate("auth-oauth-google") {
            route("/google") {
                get("/login") {}

                get("/callback") {
                    val principal: OAuthAccessTokenResponse.OAuth2? = call.principal()

                    if (principal == null) {
                        redirectAuthError(call, config.notFoundPage, "failed to get token from google")
                        return@get
                    }

                    val userInfo =
                        try {
                            OAuthHttpClient.client
                                .get(config.googleUserInfoUrl) {
                                    header(HttpHeaders.Authorization, "Bearer ${principal.accessToken}")
                                }.body<GoogleUserInfo>()
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to fetch google user info", e)
                            return@get
                        }

                    val user =
                        try {
                            userRepository.upsertUser(
                                UserUpsertInput(
                                    provider = UserProvider.GOOGLE,
                                    providerId = userInfo.sub,
                                    displayName = userInfo.name,
                                    profilePicture = userInfo.picture,
                                ),
                            )
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to authorize", e)
                            return@get
                        }

                    val token =
                        try {
                            issueToken(user.id, config)
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to issue token", e)
                            return@get
                        }

                    call.respondRedirect(
                        URLBuilder(config.frontendRedirectUrl)
                            .apply {
                                parameters.append("token", token)
                            }.build(),
                    )
                }
            }
        }

        authenticate("auth-oauth-github") {
            route("/github") {
                get("/login") {}

                get("/callback") {
                    val principal: OAuthAccessTokenResponse.OAuth2? = call.principal()
                    if (principal == null) {
                        redirectAuthError(call, config.notFoundPage, "failed to get token from github")
                        return@get
                    }

                    val userInfo =
                        try {
                            OAuthHttpClient.client
                                .get(config.githubUserInfoUrl) {
                                    header(HttpHeaders.Authorization, "Bearer ${principal.accessToken}")
                                    header(HttpHeaders.Accept, "application/vnd.github+json")
                                }.body<GithubUserInfo>()
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to fetch github user info", e)
                            return@get
                        }

                    val user =
                        try {
                            userRepository.upsertUser(
                                UserUpsertInput(
                                    provider = UserProvider.GITHUB,
                                    providerId = userInfo.id.toString(),
                                    displayName = userInfo.name,
                                    profilePicture = userInfo.avatarUrl,
                                ),
                            )
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to authorize", e)
                            return@get
                        }

                    val token =
                        try {
                            issueToken(user.id, config)
                        } catch (e: Exception) {
                            redirectAuthError(call, config.notFoundPage, "failed to issue token", e)
                            return@get
                        }

                    call.respondRedirect(
                        URLBuilder(config.frontendRedirectUrl)
                            .apply {
                                parameters.append("token", token)
                            }.build(),
                    )
                }
            }
        }
    }

    authenticate("auth-jwt") {
        route("/me") {
            get("") {
                val userId = call.requireUserId()
                val user =
                    userRepository.getUserById(userId)
                        ?: throw NotFoundException("User not found")

                call.respond(UserResponseDto.fromDomain(user))
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

fun issueToken(
    userId: Int,
    config: Config,
): String =
    JWT
        .create()
        .withAudience(config.jwtAudience)
        .withIssuer(config.jwtDomain)
        .withClaim("userId", userId)
        .withExpiresAt(Date(System.currentTimeMillis() + config.jwtExpiry * 60L * 60L * 1000L))
        .sign(Algorithm.HMAC256(config.jwtSecret))

@Serializable
data class GoogleUserInfo(
    val sub: String,
    val name: String? = null,
    val picture: String? = null,
)

@Serializable
data class GithubUserInfo(
    val id: Long,
    val name: String? = null,
    @kotlinx.serialization.SerialName("avatar_url")
    val avatarUrl: String? = null,
)
