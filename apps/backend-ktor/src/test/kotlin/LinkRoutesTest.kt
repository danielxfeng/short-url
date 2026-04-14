package dev.danielslab.shorturl.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.domain.Link
import dev.danielslab.shorturl.plugins.configureRequestValidation
import dev.danielslab.shorturl.plugins.configureSecurity
import dev.danielslab.shorturl.plugins.configureSerialization
import dev.danielslab.shorturl.plugins.configureStatusPages
import dev.danielslab.shorturl.repository.core.DuplicatedException
import dev.danielslab.shorturl.repository.core.LinkCreateInput
import dev.danielslab.shorturl.repository.core.LinkDeleteInput
import dev.danielslab.shorturl.repository.core.LinkListInput
import dev.danielslab.shorturl.repository.core.LinkRepository
import dev.danielslab.shorturl.repository.core.NotFoundException
import io.ktor.client.request.bearerAuth
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.put
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.server.routing.route
import io.ktor.server.routing.routing
import io.ktor.server.testing.testApplication
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.withTimeout
import java.util.Date
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Instant

class LinkRoutesTest {
    @Test
    fun `get code redirects to original url and records click asynchronously`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.linksByCode["hello-world"] =
                Link(
                    id = 1,
                    userId = 7,
                    code = "hello-world",
                    originalUrl = "https://example.com/page",
                    clicks = 2,
                    note = "note",
                    createdAt = Instant.parse("2024-01-01T00:00:00Z"),
                )

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/short-urls/hello-world").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("https://example.com/page", headers["Location"])
            }

            withTimeout(1_000.milliseconds) {
                assertEquals("hello-world", repository.clickedCode.await())
            }
        }

    @Test
    fun `get code redirects invalid format to invalid url page`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/short-urls/hello_123").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("http://localhost:5173/not-found?invalid-url=hello_123", headers["Location"])
            }
        }

    @Test
    fun `get code redirects missing link to invalid url page`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/short-urls/missing").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("http://localhost:5173/not-found?invalid-url=missing", headers["Location"])
            }
        }

    @Test
    fun `get code redirects repository error to invalid url page`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failGetCode = "broken"

            application {
                configureSecurity(config)
                configureSerialization()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            createClient {
                followRedirects = false
            }.get("/api/v1/short-urls/broken").apply {
                assertEquals(HttpStatusCode.Found, status)
                assertEquals("http://localhost:5173/not-found?invalid-url=broken", headers["Location"])
            }
        }

    @Test
    fun `list short urls returns unauthorized without token`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client.get("/api/v1/short-urls").apply {
                assertEquals(HttpStatusCode.Unauthorized, status)
            }
        }

    @Test
    fun `list short urls returns links with default pagination`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.linksToList =
                listOf(
                    Link(
                        id = 2,
                        userId = 7,
                        code = "two",
                        originalUrl = "https://example.com/two",
                        clicks = 5,
                        note = null,
                        createdAt = Instant.parse("2024-01-02T00:00:00Z"),
                    ),
                    Link(
                        id = 1,
                        userId = 7,
                        code = "one",
                        originalUrl = "https://example.com/one",
                        clicks = 0,
                        note = "first",
                        createdAt = Instant.parse("2024-01-01T00:00:00Z"),
                        deletedAt = Instant.parse("2024-01-03T00:00:00Z"),
                    ),
                )

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/short-urls") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.OK, status)
                    assertEquals(
                        """[{"id":2,"userId":7,"code":"two","originalUrl":"https://example.com/two","clicks":5,"note":null,"createdAt":"2024-01-02T00:00:00Z","isDeleted":false},{"id":1,"userId":7,"code":"one","originalUrl":"https://example.com/one","clicks":0,"note":"first","createdAt":"2024-01-01T00:00:00Z","isDeleted":true}]""",
                        bodyAsText(),
                    )
                }

            assertEquals(LinkListInput(userId = 7, cursor = null, limit = 20), repository.lastListInput)
        }

    @Test
    fun `list short urls uses cursor and clamps limit`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/short-urls?cursor=123&limit=999") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.OK, status)
                    assertEquals("[]", bodyAsText())
                }

            assertEquals(LinkListInput(userId = 7, cursor = "123", limit = 100), repository.lastListInput)
        }

    @Test
    fun `list short urls normalizes invalid limit to default`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/short-urls?limit=abc") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.OK, status)
                }

            assertEquals(LinkListInput(userId = 7, cursor = null, limit = 20), repository.lastListInput)
        }

    @Test
    fun `list short urls rejects invalid cursor`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .get("/api/v1/short-urls?cursor=") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.BadRequest, status)
                    assertEquals("""{"error":"cursor must not be blank, and at most 255 characters"}""", bodyAsText())
                }
        }

    @Test
    fun `create short url returns unauthorized without token`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client.post("/api/v1/short-urls").apply {
                assertEquals(HttpStatusCode.Unauthorized, status)
            }
        }

    @Test
    fun `create short url returns created with json body`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.createdLink =
                Link(
                    id = 1,
                    userId = 7,
                    code = "hello-world",
                    originalUrl = "https://example.com/page",
                    clicks = 0,
                    note = "my note",
                    createdAt = Instant.parse("2024-01-01T00:00:00Z"),
                )

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .post("/api/v1/short-urls") {
                    bearerAuth(issueTestToken(7, config))
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":"hello-world","originalUrl":"https://example.com/page","note":"my note"}""")
                }.apply {
                    assertEquals(HttpStatusCode.Created, status)
                    assertEquals(
                        """{"id":1,"userId":7,"code":"hello-world","originalUrl":"https://example.com/page","clicks":0,"note":"my note","createdAt":"2024-01-01T00:00:00Z","isDeleted":false}""",
                        bodyAsText(),
                    )
                }

            assertEquals(
                LinkCreateInput(
                    userId = 7,
                    code = "hello-world",
                    originalUrl = "https://example.com/page",
                    note = "my note",
                ),
                repository.lastCreateInput,
            )
        }

    @Test
    fun `create short url generates code when omitted`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .post("/api/v1/short-urls") {
                    bearerAuth(issueTestToken(7, config))
                    contentType(ContentType.Application.Json)
                    setBody("""{"originalUrl":"https://example.com/page","note":"my note"}""")
                }.apply {
                    assertEquals(HttpStatusCode.Created, status)
                }

            val createdInput = repository.lastCreateInput
            assertEquals(7, createdInput?.userId)
            assertEquals("https://example.com/page", createdInput?.originalUrl)
            assertEquals("my note", createdInput?.note)
            assertEquals(config.linkCodeLength, createdInput?.code?.length)
            assertEquals(true, createdInput?.code?.all { it.isLetterOrDigit() || it == '-' })
        }

    @Test
    fun `create short url returns bad request for invalid body`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .post("/api/v1/short-urls") {
                    bearerAuth(issueTestToken(7, config))
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":"","originalUrl":"https://example.com/page"}""")
                }.apply {
                    assertEquals(HttpStatusCode.BadRequest, status)
                    assertEquals(
                        """{"error":"Validation failed for LinkCreateRequestDto(code=, originalUrl=https://example.com/page, note=null). Reasons: code must not be blank"}""",
                        bodyAsText(),
                    )
                }
        }

    @Test
    fun `create short url returns conflict for duplicate code`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failCreateDuplicate = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .post("/api/v1/short-urls") {
                    bearerAuth(issueTestToken(7, config))
                    contentType(ContentType.Application.Json)
                    setBody("""{"code":"hello-world","originalUrl":"https://example.com/page"}""")
                }.apply {
                    assertEquals(HttpStatusCode.Conflict, status)
                    assertEquals("""{"error":"Link already exists"}""", bodyAsText())
                }
        }

    @Test
    fun `restore short url returns no content`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .put("/api/v1/short-urls/restore-me/restore") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NoContent, status)
                }

            assertEquals(LinkDeleteInput(userId = 7, code = "restore-me"), repository.lastRestoreInput)
        }

    @Test
    fun `restore short url returns bad request for invalid code`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .put("/api/v1/short-urls/hello_123/restore") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.BadRequest, status)
                    assertEquals("""{"error":"code must contain only alphabets, numbers, and dash"}""", bodyAsText())
                }
        }

    @Test
    fun `restore short url returns not found when link does not exist`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failRestoreMissing = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .put("/api/v1/short-urls/missing/restore") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
                }
        }

    @Test
    fun `restore short url returns not found when link is not soft deleted`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failRestoreActive = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .put("/api/v1/short-urls/active-link/restore") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
                }
        }

    @Test
    fun `delete short url returns no content`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/delete-me") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NoContent, status)
                }

            assertEquals(LinkDeleteInput(userId = 7, code = "delete-me"), repository.lastSoftDeleteInput)
        }

    @Test
    fun `delete short url returns not found when link does not exist`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failSoftDeleteMissing = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/missing") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
                }
        }

    @Test
    fun `delete short url returns not found when link is already soft deleted`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failSoftDeleteAlreadyDeleted = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/deleted-link") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
                }
        }

    @Test
    fun `permanent delete short url returns no content`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/delete-me/permanent") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NoContent, status)
                }

            assertEquals(LinkDeleteInput(userId = 7, code = "delete-me"), repository.lastPermanentDeleteInput)
        }

    @Test
    fun `permanent delete short url returns not found when link does not exist`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failPermanentDeleteMissing = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/missing/permanent") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
                }
        }

    @Test
    fun `permanent delete short url returns not found when link is not soft deleted`() =
        testApplication {
            val config = testConfig()
            val repository = FakeLinkRepository()
            repository.failPermanentDeleteActive = true

            application {
                configureSecurity(config)
                configureSerialization()
                configureRequestValidation()
                configureStatusPages()
                routing {
                    route("/api/v1/short-urls") {
                        linkRoutes(config, repository)
                    }
                }
            }

            client
                .delete("/api/v1/short-urls/active-link/permanent") {
                    bearerAuth(issueTestToken(7, config))
                }.apply {
                    assertEquals(HttpStatusCode.NotFound, status)
                    assertEquals("""{"error":"Link not found"}""", bodyAsText())
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

private class FakeLinkRepository : LinkRepository {
    val linksByCode = mutableMapOf<String, Link>()
    var linksToList: List<Link> = emptyList()
    var createdLink: Link? = null

    var lastCreateInput: LinkCreateInput? = null
    var lastListInput: LinkListInput? = null
    var lastRestoreInput: LinkDeleteInput? = null
    var lastSoftDeleteInput: LinkDeleteInput? = null
    var lastPermanentDeleteInput: LinkDeleteInput? = null

    var failGetCode: String? = null
    var failCreateDuplicate = false
    var failRestoreMissing = false
    var failRestoreActive = false
    var failSoftDeleteMissing = false
    var failSoftDeleteAlreadyDeleted = false
    var failPermanentDeleteMissing = false
    var failPermanentDeleteActive = false

    val clickedCode = CompletableDeferred<String>()

    override suspend fun createLink(request: LinkCreateInput): Link {
        lastCreateInput = request
        if (failCreateDuplicate) {
            throw DuplicatedException("Link already exists")
        }
        return createdLink
            ?: Link(
                id = 1,
                userId = request.userId,
                code = request.code,
                originalUrl = request.originalUrl,
                clicks = 0,
                note = request.note,
                createdAt = Instant.parse("2024-01-01T00:00:00Z"),
            )
    }

    override suspend fun getLinkByCode(code: String): Link? {
        if (failGetCode == code) {
            throw IllegalStateException("boom")
        }
        return linksByCode[code]
    }

    override suspend fun getLinkByCodeWithDeleted(code: String): Link? = linksByCode[code]

    override suspend fun getLinksByUserId(request: LinkListInput): List<Link> {
        lastListInput = request
        return linksToList
    }

    override suspend fun softDeleteLinkById(request: LinkDeleteInput) {
        lastSoftDeleteInput = request
        if (failSoftDeleteMissing || failSoftDeleteAlreadyDeleted) {
            throw NotFoundException("Link not found")
        }
    }

    override suspend fun restoreLinkById(request: LinkDeleteInput) {
        lastRestoreInput = request
        if (failRestoreMissing || failRestoreActive) {
            throw NotFoundException("Link not found")
        }
    }

    override suspend fun permanentlyDeleteLinkById(request: LinkDeleteInput) {
        lastPermanentDeleteInput = request
        if (failPermanentDeleteMissing || failPermanentDeleteActive) {
            throw NotFoundException("Link not found")
        }
    }

    override suspend fun setLinkClicked(code: String): Int {
        if (!clickedCode.isCompleted) {
            clickedCode.complete(code)
        }
        return 1
    }
}
