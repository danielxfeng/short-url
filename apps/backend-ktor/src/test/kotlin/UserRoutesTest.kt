package dev.danielslab.shorturl.routes

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.plugins.configureSecurity
import dev.danielslab.shorturl.plugins.configureSerialization
import dev.danielslab.shorturl.plugins.configureStatusPages
import dev.danielslab.shorturl.repository.core.UserRepository
import dev.danielslab.shorturl.repository.core.UserUpsertInput
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.routing.*
import io.ktor.server.testing.*
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class UserRoutesTest {

    @Test
    fun `get me returns unauthorized without token`() = testApplication {
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
    fun `get me returns current user`() = testApplication {
        val config = testConfig()
        val repository = FakeUserRepository()
        repository.users[1] = User(
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

        client.get("/api/v1/user/me") {
            bearerAuth(issueToken(1, config))
        }.apply {
            assertEquals(HttpStatusCode.OK, status)
            assertEquals(
                """{"id":1,"provider":"GOOGLE","providerId":"google-1","displayName":"Daniel","profilePicture":"https://example.com/a.png"}""",
                bodyAsText(),
            )
        }
    }

    @Test
    fun `get me returns not found when user is missing`() = testApplication {
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

        client.get("/api/v1/user/me") {
            bearerAuth(issueToken(1, config))
        }.apply {
            assertEquals(HttpStatusCode.NotFound, status)
        }
    }

    @Test
    fun `delete me deletes current user`() = testApplication {
        val config = testConfig()
        val repository = FakeUserRepository()
        repository.users[1] = User(
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

        client.delete("/api/v1/user/me") {
            bearerAuth(issueToken(1, config))
        }.apply {
            assertEquals(HttpStatusCode.NoContent, status)
        }

        assertNull(repository.users[1])
    }
}

private fun testConfig(): Config = Config(
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
)

private class FakeUserRepository : UserRepository {
    val users = mutableMapOf<Int, User>()
    private var nextId = 1

    override suspend fun getUserById(id: Int): User? = users[id]

    override suspend fun deleteUserById(id: Int) {
        users.remove(id)
    }

    override suspend fun upsertUser(userInput: UserUpsertInput): User {
        val existing = users.values.firstOrNull {
            it.provider == userInput.provider && it.providerId == userInput.providerId
        }

        val user = if (existing != null) {
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
