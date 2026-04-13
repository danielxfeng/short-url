package dev.danielslab.shorturl

import dev.danielslab.shorturl.db.exposed.DatabaseFactory
import dev.danielslab.shorturl.db.exposed.DbExecutor
import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.repository.core.DuplicatedException
import dev.danielslab.shorturl.repository.core.LinkCreateInput
import dev.danielslab.shorturl.repository.core.LinkDeleteInput
import dev.danielslab.shorturl.repository.core.LinkListInput
import dev.danielslab.shorturl.repository.core.NotFoundException
import dev.danielslab.shorturl.repository.core.UserUpsertInput
import dev.danielslab.shorturl.repository.pq.PqRepository
import io.github.cdimascio.dotenv.Dotenv
import kotlinx.coroutines.runBlocking
import org.jetbrains.exposed.sql.transactions.TransactionManager
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class PqRepositoryTest {
    private val testContext = TestContext

    @BeforeTest
    fun resetDb() = runBlocking {
        testContext.resetDb()
    }

    @Test
    fun `getUserById returns null when user is missing`() = runBlocking {
        assertNull(testContext.repository.getUserById(1))
    }

    @Test
    fun `upsertUser inserts when user does not exist`() = runBlocking {
        testContext.repository.upsertUser(
            UserUpsertInput(
                provider = UserProvider.GOOGLE,
                providerId = "google-1",
                displayName = "Daniel",
                profilePicture = "https://example.com/a.png",
            )
        )

        val user = testContext.repository.getUserById(1)
        assertNotNull(user)
        assertEquals(1, user.id)
        assertEquals(UserProvider.GOOGLE, user.provider)
        assertEquals("google-1", user.providerId)
        assertEquals("Daniel", user.displayName)
        assertEquals("https://example.com/a.png", user.profilePicture)
        assertEquals(1, testContext.countUsers())
    }

    @Test
    fun `upsertUser updates existing user without creating duplicate row`() = runBlocking {
        seedUser(
            provider = UserProvider.GOOGLE,
            providerId = "google-1",
            displayName = "Before",
            profilePicture = "https://example.com/a.png",
        )

        testContext.repository.upsertUser(
            UserUpsertInput(
                provider = UserProvider.GOOGLE,
                providerId = "google-1",
                displayName = "After",
                profilePicture = "https://example.com/b.png",
            )
        )

        val user = testContext.repository.getUserById(1)
        assertNotNull(user)
        assertEquals(1, user.id)
        assertEquals("After", user.displayName)
        assertEquals("https://example.com/b.png", user.profilePicture)
        assertEquals(1, testContext.countUsers())
    }

    @Test
    fun `deleteUserById deletes existing user`() = runBlocking {
        seedUser()

        testContext.repository.deleteUserById(1)

        assertNull(testContext.repository.getUserById(1))
        assertEquals(0, testContext.countUsers())
    }

    @Test
    fun `deleteUserById throws not found when user is missing`() {
        runBlocking {
            assertFailsWith<NotFoundException> {
                testContext.repository.deleteUserById(1)
            }
        }
    }

    @Test
    fun `createLink creates active link with note`() = runBlocking {
        seedUser()

        val link = testContext.repository.createLink(
            LinkCreateInput(
                userId = 1,
                code = "with-note",
                originalUrl = "https://example.com/with-note",
                note = "note",
            )
        )

        assertEquals(1, link.id)
        assertEquals(1, link.userId)
        assertEquals("with-note", link.code)
        assertEquals("https://example.com/with-note", link.originalUrl)
        assertEquals("note", link.note)
        assertEquals(0, link.clicks)
        assertNull(link.deletedAt)
    }

    @Test
    fun `createLink creates active link without note`() = runBlocking {
        seedUser()

        val link = testContext.repository.createLink(
            LinkCreateInput(
                userId = 1,
                code = "without-note",
                originalUrl = "https://example.com/without-note",
                note = null,
            )
        )

        assertEquals("without-note", link.code)
        assertNull(link.note)
        assertEquals(0, link.clicks)
        assertNull(link.deletedAt)
    }

    @Test
    fun `createLink throws duplicated when code already exists`() {
        runBlocking {
            seedUser()
            testContext.repository.createLink(
                LinkCreateInput(
                    userId = 1,
                    code = "dup",
                    originalUrl = "https://example.com/a",
                )
            )

            assertFailsWith<DuplicatedException> {
                testContext.repository.createLink(
                    LinkCreateInput(
                        userId = 1,
                        code = "dup",
                        originalUrl = "https://example.com/b",
                    )
                )
            }
        }
    }

    @Test
    fun `getLinkByCode returns active link and null when missing`() = runBlocking {
        seedUser()
        val created = seedLink(code = "hello")

        val found = testContext.repository.getLinkByCode("hello")

        assertNotNull(found)
        assertEquals(created.id, found.id)
        assertNull(testContext.repository.getLinkByCode("missing"))
    }

    @Test
    fun `getLinkByCode returns null when link is soft deleted and getLinkByCodeWithDeleted still returns it`() {
        runBlocking {
            seedUser()
            seedLink(code = "gone")
            testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "gone"))

            assertNull(testContext.repository.getLinkByCode("gone"))

            val withDeleted = testContext.repository.getLinkByCodeWithDeleted("gone")
            assertNotNull(withDeleted)
            assertEquals("gone", withDeleted.code)
            assertNotNull(withDeleted.deletedAt)
        }
    }

    @Test
    fun `getLinksByUserId returns newest first, respects limit, and paginates by id cursor`() = runBlocking {
        seedUser()
        repeat(5) { index ->
            seedLink(code = "code-${index + 1}", originalUrl = "https://example.com/${index + 1}")
        }

        val firstPage = testContext.repository.getLinksByUserId(
            LinkListInput(userId = 1, cursor = null, limit = 2)
        )
        assertEquals(listOf(5, 4), firstPage.map { it.id })

        val secondPage = testContext.repository.getLinksByUserId(
            LinkListInput(userId = 1, cursor = firstPage.last().id.toString(), limit = 2)
        )
        assertEquals(listOf(3, 2), secondPage.map { it.id })

        val thirdPage = testContext.repository.getLinksByUserId(
            LinkListInput(userId = 1, cursor = secondPage.last().id.toString(), limit = 2)
        )
        assertEquals(listOf(1), thirdPage.map { it.id })
    }

    @Test
    fun `getLinksByUserId includes soft deleted rows and only returns current user rows`() {
        runBlocking {
            seedUser(providerId = "google-1")
            testContext.repository.upsertUser(
                UserUpsertInput(
                    provider = UserProvider.GITHUB,
                    providerId = "github-2",
                    displayName = "Other",
                    profilePicture = null,
                )
            )

            seedLink(userId = 1, code = "active-1")
            seedLink(userId = 1, code = "deleted-1")
            seedLink(userId = 2, code = "other-user")
            testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "deleted-1"))

            val links = testContext.repository.getLinksByUserId(
                LinkListInput(userId = 1, cursor = null, limit = 10)
            )

            assertEquals(listOf("deleted-1", "active-1"), links.map { it.code })
            assertNotNull(links.first().deletedAt)
        }
    }

    @Test
    fun `getLinksByUserId returns empty list when user has no links`() = runBlocking {
        seedUser()

        val links = testContext.repository.getLinksByUserId(
            LinkListInput(userId = 1, cursor = null, limit = 10)
        )

        assertEquals(emptyList(), links)
    }

    @Test
    fun `softDeleteLinkById soft deletes active link`() {
        runBlocking {
            seedUser()
            seedLink(code = "delete-me")

            testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "delete-me"))

            val link = testContext.repository.getLinkByCodeWithDeleted("delete-me")
            assertNotNull(link)
            assertNotNull(link.deletedAt)
        }
    }

    @Test
    fun `softDeleteLinkById throws not found when link is missing or already deleted`() {
        runBlocking {
            seedUser()
            seedLink(code = "already-deleted")
            testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "already-deleted"))

            assertFailsWith<NotFoundException> {
                testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "missing"))
            }
            assertFailsWith<NotFoundException> {
                testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "already-deleted"))
            }
        }
    }

    @Test
    fun `restoreLinkById restores deleted link`() = runBlocking {
        seedUser()
        seedLink(code = "restore-me")
        testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "restore-me"))

        testContext.repository.restoreLinkById(LinkDeleteInput(userId = 1, code = "restore-me"))

        val link = testContext.repository.getLinkByCode("restore-me")
        assertNotNull(link)
        assertNull(link.deletedAt)
    }

    @Test
    fun `restoreLinkById throws not found when link is missing or active`() {
        runBlocking {
            seedUser()
            seedLink(code = "active-link")

            assertFailsWith<NotFoundException> {
                testContext.repository.restoreLinkById(LinkDeleteInput(userId = 1, code = "missing"))
            }
            assertFailsWith<NotFoundException> {
                testContext.repository.restoreLinkById(LinkDeleteInput(userId = 1, code = "active-link"))
            }
        }
    }

    @Test
    fun `permanentlyDeleteLinkById deletes deleted link only`() = runBlocking {
        seedUser()
        seedLink(code = "hard-delete")
        testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "hard-delete"))

        testContext.repository.permanentlyDeleteLinkById(LinkDeleteInput(userId = 1, code = "hard-delete"))

        assertNull(testContext.repository.getLinkByCodeWithDeleted("hard-delete"))
    }

    @Test
    fun `permanentlyDeleteLinkById throws not found when link is missing or active`() {
        runBlocking {
            seedUser()
            seedLink(code = "active-link")

            assertFailsWith<NotFoundException> {
                testContext.repository.permanentlyDeleteLinkById(LinkDeleteInput(userId = 1, code = "missing"))
            }
            assertFailsWith<NotFoundException> {
                testContext.repository.permanentlyDeleteLinkById(LinkDeleteInput(userId = 1, code = "active-link"))
            }
        }
    }

    @Test
    fun `setLinkClicked increments active link and returns updated count`() = runBlocking {
        seedUser()
        seedLink(code = "click-me")

        assertEquals(1, testContext.repository.setLinkClicked("click-me"))
        assertEquals(2, testContext.repository.setLinkClicked("click-me"))
    }

    @Test
    fun `setLinkClicked throws not found for missing or soft deleted link`() {
        runBlocking {
            seedUser()
            seedLink(code = "deleted-click")
            testContext.repository.softDeleteLinkById(LinkDeleteInput(userId = 1, code = "deleted-click"))

            assertFailsWith<NotFoundException> {
                testContext.repository.setLinkClicked("missing")
            }
            assertFailsWith<NotFoundException> {
                testContext.repository.setLinkClicked("deleted-click")
            }
        }
    }

    private suspend fun seedUser(
        provider: UserProvider = UserProvider.GOOGLE,
        providerId: String = "google-1",
        displayName: String? = "Daniel",
        profilePicture: String? = null,
    ) {
        testContext.repository.upsertUser(
            UserUpsertInput(
                provider = provider,
                providerId = providerId,
                displayName = displayName,
                profilePicture = profilePicture,
            )
        )
    }

    private suspend fun seedLink(
        userId: Int = 1,
        code: String,
        originalUrl: String = "https://example.com/$code",
        note: String? = null,
    ) = testContext.repository.createLink(
        LinkCreateInput(
            userId = userId,
            code = code,
            originalUrl = originalUrl,
            note = note,
        )
    )

    private object TestContext {
        private val dotenv: Dotenv = Dotenv.configure()
            .ignoreIfMissing()
            .load()
        private val database = DatabaseFactory().connect(testDbUrl())
        private val dbExecutor = DbExecutor(database)

        val repository = PqRepository(dbExecutor)

        suspend fun resetDb() {
            dbExecutor.query {
                TransactionManager.current().exec("TRUNCATE users, links RESTART IDENTITY CASCADE;")
            }
        }

        suspend fun countUsers(): Int = dbExecutor.query {
            TransactionManager.current().exec("SELECT COUNT(*) FROM users") { resultSet ->
                if (resultSet.next()) resultSet.getInt(1) else 0
            } ?: 0
        }

        private fun testDbUrl(): String =
            System.getenv("TEST_DB_URL")
                ?: dotenv["TEST_DB_URL"]
                ?: error("TEST_DB_URL is missing")
    }
}
