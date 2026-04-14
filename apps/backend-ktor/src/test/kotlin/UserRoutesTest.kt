package dev.danielslab.shorturl.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.plugins.configureSecurity
import dev.danielslab.shorturl.plugins.configureSerialization
import dev.danielslab.shorturl.plugins.configureStatusPages
import dev.danielslab.shorturl.repository.core.NotFoundException
import dev.danielslab.shorturl.repository.core.UserRepository
import dev.danielslab.shorturl.repository.core.UserUpsertInput
import io.ktor.client.request.bearerAuth
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.server.routing.route
import io.ktor.server.routing.routing
import io.ktor.server.testing.testApplication
import java.util.Date
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class UserRoutesTest {
    @Test
    fun `auth provider redirects to google login path`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/user/auth/google").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("/api/v1/user/auth/google/login", headers["Location"])
            }
        }

    @Test
    fun `auth provider redirects invalid provider to not found page`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/user/auth/discord").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("http://localhost:5173/auth/callback?error=provider+not+found", headers["Location"])
            }
        }

    @Test
    fun `get me returns unauthorized without token`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client.get("/api/v1/user/me").apply {
                assertEquals(HttpStatusCode.Unauthorized, status)
            }
        }

    @Test
    fun `get me returns current user`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()
            repository.users[1] =
                User(
                    id = 1,
                    provider = UserProvider.GOOGLE,
                    providerId = "google-1",
                    displayName = "Daniel",
                    profilePicture = "https://example.com/a.png",
                )

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/user/me") {
                    bearerAuth(issueTestToken(1, config))
                }.apply {
                    assertEquals(HttpStatusCode.OK, status)
                    assertEquals(
                        """{"id":1,"provider":"GOOGLE","providerId":"google-1","displayName":"Daniel","profilePicture":"https://example.com/a.png"}""",
                        bodyAsText(),
                    )
                }
        }

    @Test
    fun `get me returns not found when user is missing`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/user/me") {
                    bearerAuth(issueTestToken(1, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"User not found"}""", bodyAsText())
                }
        }

    @Test
    fun `delete me returns unauthorized without token`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client.delete("/api/v1/user/me").apply {
                assertEquals(HttpStatusCode.Unauthorized, status)
            }
        }

    @Test
    fun `delete me deletes current user`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()
            repository.users[1] =
                User(
                    id = 1,
                    provider = UserProvider.GITHUB,
                    providerId = "123",
                    displayName = "Daniel",
                    profilePicture = null,
                )

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/user/me") {
                    bearerAuth(issueTestToken(1, config))
                }.apply {
                    assertEquals(HttpStatusCode.NoContent, status)
                }

            assertNull(repository.users[1])
        }

    @Test
    fun `delete me returns not found when user is missing`() =
        testApplication {
            val config = testConfig()
            val repository = FakeUserRepository()
            repository.failDeleteMissing = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/user") {
                        userRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/user/me") {
                    bearerAuth(issueTestToken(1, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"User not found"}""", bodyAsText())
                }
        }
}

private fun testConfig(): Config =
    Config(
        port = 8081,
        cors = "http://localhost:5173",
        backendPublicUrl = "http://localhost:8081",
        dbUrl = "jdbc:postgresql://localhost/test",
        testDbUrl = "jdbc:postgresql://localhost/test",
        jwtDomain = "https://api-s.danielslab.dev",
        jwtAudience = "danielslab-short-url",
        jwtRealm = "short-url-ktor",
        jwtSecret = "test-secret",
        jwtExpiry = 1,
        notFoundPage = "http://localhost:5173/not-found",
        googleClientId = "google-client-id",
        googleClientSecret = "google-client-secret",
        githubClientId = "github-client-id",
        githubClientSecret = "github-client-secret",
        googleAuthorizeUrl = "https://accounts.google.com/o/oauth2/auth",
        googleAccessTokenUrl = "https://accounts.google.com/o/oauth2/token",
        googleUserInfoUrl = "https://openidconnect.googleapis.com/v1/userinfo",
        googleScopes = "openid,profile",
        githubAuthorizeUrl = "https://github.com/login/oauth/authorize",
        githubAccessTokenUrl = "https://github.com/login/oauth/access_token",
        githubUserInfoUrl = "https://api.github.com/user",
        githubScopes = "read:user",
        frontendRedirectUrl = "http://localhost:5173/auth/callback",
        linkDefaultPageSize = 20,
        linkMaxPageSize = 100,
        linkCodeLength = 8,
        linkCodeGenerateRetries = 5,
    )

private fun issueTestToken(
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

private class FakeUserRepository : UserRepository {
    val users = mutableMapOf<Int, User>()
    private var nextId = 1
    var failDeleteMissing = false

    override suspend fun getUserById(id: Int): User? = users[id]

    override suspend fun deleteUserById(id: Int) {
        val removed = users.remove(id)
        if (removed == null && failDeleteMissing) {
            throw NotFoundException("User not found")
        }
    }

    override suspend fun upsertUser(userInput: UserUpsertInput): User {
        val existing =
            users.values.firstOrNull {
                it.provider == userInput.provider && it.providerId == userInput.providerId
            }

        val user =
            if (existing != null) {
                existing.copy(
                    displayName = userInput.displayName,
                    profilePicture = userInput.profilePicture,
                )
            } else {
                User(
                    id = nextId++,
                    provider = userInput.provider,
                    providerId = userInput.providerId,
                    displayName = userInput.displayName,
                    profilePicture = userInput.profilePicture,
                )
            }

        users[user.id] = user
        return user
    }
}
