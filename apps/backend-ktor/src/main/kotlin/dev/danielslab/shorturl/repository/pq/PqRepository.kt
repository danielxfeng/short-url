package dev.danielslab.shorturl.repository.pq

import dev.danielslab.shorturl.db.exposed.DbExecutor
import dev.danielslab.shorturl.db.schema.LinksTable
import dev.danielslab.shorturl.db.schema.UsersTable
import dev.danielslab.shorturl.domain.Link
import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.repository.core.DuplicatedException
import dev.danielslab.shorturl.repository.core.LinkCreateInput
import dev.danielslab.shorturl.repository.core.LinkDeleteInput
import dev.danielslab.shorturl.repository.core.LinkListInput
import dev.danielslab.shorturl.repository.core.LinkRepository
import dev.danielslab.shorturl.repository.core.NotFoundException
import dev.danielslab.shorturl.repository.core.UserRepository
import dev.danielslab.shorturl.repository.core.UserUpsertInput
import org.jetbrains.exposed.exceptions.ExposedSQLException
import org.jetbrains.exposed.sql.ResultRow
import org.jetbrains.exposed.sql.SortOrder
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.SqlExpressionBuilder.isNotNull
import org.jetbrains.exposed.sql.SqlExpressionBuilder.isNull
import org.jetbrains.exposed.sql.SqlExpressionBuilder.plus
import org.jetbrains.exposed.sql.and
import org.jetbrains.exposed.sql.andWhere
import org.jetbrains.exposed.sql.deleteWhere
import org.jetbrains.exposed.sql.insertReturning
import org.jetbrains.exposed.sql.selectAll
import org.jetbrains.exposed.sql.update
import org.jetbrains.exposed.sql.updateReturning
import org.jetbrains.exposed.sql.upsert
import org.postgresql.util.PSQLException
import kotlin.time.toKotlinInstant

class PqRepository(
    private val dbExecutor: DbExecutor,
) : UserRepository,
    LinkRepository {
    override suspend fun getUserById(id: Int): User? =
        dbExecutor.query {
            UsersTable
                .selectAll()
                .where { UsersTable.id eq id }
                .singleOrNull()
                ?.toUser()
        }

    override suspend fun deleteUserById(id: Int): Unit =
        dbExecutor.query {
            val deletedRows =
                UsersTable
                    .deleteWhere { UsersTable.id eq id }

            if (deletedRows == 0) {
                throw NotFoundException("User not found")
            }
        }

    override suspend fun upsertUser(userInput: UserUpsertInput): User =
        dbExecutor.query {
            UsersTable.upsert(
                keys = arrayOf(UsersTable.provider, UsersTable.providerId),
            ) {
                it[provider] = userInput.provider
                it[providerId] = userInput.providerId
                it[displayName] = userInput.displayName
                it[profilePic] = userInput.profilePicture
            }

            UsersTable
                .selectAll()
                .where {
                    (UsersTable.provider eq userInput.provider) and
                        (UsersTable.providerId eq userInput.providerId)
                }.single()
                .toUser()
        }

    override suspend fun createLink(request: LinkCreateInput): Link =
        try {
            dbExecutor.query {
                LinksTable
                    .insertReturning {
                        it[userId] = request.userId
                        it[code] = request.code
                        it[originalUrl] = request.originalUrl
                        it[note] = request.note
                    }.single()
                    .toLink()
            }
        } catch (error: ExposedSQLException) {
            if (isUniqueViolation(error)) {
                throw DuplicatedException("Link already exists")
            }
            throw error
        }

    override suspend fun getLinkByCode(code: String): Link? =
        dbExecutor.query {
            LinksTable
                .selectAll()
                .where { (LinksTable.code eq code) and LinksTable.deletedAt.isNull() }
                .singleOrNull()
                ?.toLink()
        }

    override suspend fun getLinkByCodeWithDeleted(code: String): Link? =
        dbExecutor.query {
            LinksTable
                .selectAll()
                .where { LinksTable.code eq code }
                .singleOrNull()
                ?.toLink()
        }

    override suspend fun getLinksByUserId(request: LinkListInput): List<Link> =
        dbExecutor.query {
            LinksTable
                .selectAll()
                .where { LinksTable.userId eq request.userId }
                .apply {
                    request.cursor?.toInt()?.let { cursorId ->
                        andWhere { LinksTable.id less cursorId }
                    }
                }.orderBy(LinksTable.id to SortOrder.DESC)
                .limit(request.limit)
                .map { it.toLink() }
        }

    override suspend fun softDeleteLinkById(request: LinkDeleteInput): Unit =
        dbExecutor.query {
            val updatedRows =
                LinksTable
                    .update({
                        (LinksTable.userId eq request.userId) and (LinksTable.code eq request.code) and (LinksTable.deletedAt.isNull())
                    }) {
                        it[deletedAt] = java.time.OffsetDateTime.now(java.time.ZoneOffset.UTC)
                    }

            if (updatedRows == 0) {
                throw NotFoundException("Link not found")
            }
        }

    override suspend fun restoreLinkById(request: LinkDeleteInput): Unit =
        dbExecutor.query {
            val updatedRows =
                LinksTable.update({
                    (LinksTable.userId eq request.userId) and (LinksTable.code eq request.code) and (LinksTable.deletedAt.isNotNull())
                }) {
                    it[deletedAt] = null
                }

            if (updatedRows == 0) {
                throw NotFoundException("Link not found")
            }
        }

    override suspend fun permanentlyDeleteLinkById(request: LinkDeleteInput): Unit =
        dbExecutor.query {
            val deletedRows =
                LinksTable.deleteWhere {
                    (LinksTable.userId eq request.userId) and (LinksTable.code eq request.code) and (LinksTable.deletedAt.isNotNull())
                }

            if (deletedRows == 0) {
                throw NotFoundException("Link not found")
            }
        }

    override suspend fun setLinkClicked(code: String): Int =
        dbExecutor.query {
            LinksTable
                .updateReturning(
                    returning = listOf(LinksTable.clicks),
                    where = { (LinksTable.code eq code) and (LinksTable.deletedAt.isNull()) },
                ) {
                    it[clicks] = LinksTable.clicks + 1
                }.singleOrNull()
                ?.get(LinksTable.clicks)
                ?: throw NotFoundException("Link not found")
        }

    private fun isUniqueViolation(error: ExposedSQLException): Boolean {
        val cause = error.cause
        return cause is PSQLException && cause.sqlState == "23505"
    }

    private fun ResultRow.toUser(): User =
        User(
            id = this[UsersTable.id],
            provider = this[UsersTable.provider],
            providerId = this[UsersTable.providerId],
            displayName = this[UsersTable.displayName],
            profilePicture = this[UsersTable.profilePic],
            createdAt = this[UsersTable.createdAt].toInstant().toKotlinInstant(),
        )

    private fun ResultRow.toLink(): Link =
        Link(
            id = this[LinksTable.id],
            userId = this[LinksTable.userId],
            code = this[LinksTable.code],
            originalUrl = this[LinksTable.originalUrl],
            clicks = this[LinksTable.clicks],
            note = this[LinksTable.note],
            createdAt = this[LinksTable.createdAt].toInstant().toKotlinInstant(),
            deletedAt = this[LinksTable.deletedAt]?.toInstant()?.toKotlinInstant(),
        )
}
