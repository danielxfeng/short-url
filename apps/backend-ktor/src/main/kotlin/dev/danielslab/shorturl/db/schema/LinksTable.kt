package dev.danielslab.shorturl.db.schema

import org.jetbrains.exposed.sql.ReferenceOption
import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.javatime.CurrentTimestampWithTimeZone
import org.jetbrains.exposed.sql.javatime.timestampWithTimeZone

object LinksTable : Table("links") {
    val id = integer("id").autoIncrement()
    val userId = integer("user_id").references(UsersTable.id, onDelete = ReferenceOption.CASCADE)
    val code = text("code")
    val originalUrl = text("original_url")
    val clicks = integer("clicks").default(0)
    val note = text("note").nullable()
    val createdAt = timestampWithTimeZone("created_at").defaultExpression(CurrentTimestampWithTimeZone)
    val deletedAt = timestampWithTimeZone("deleted_at").nullable()

    override val primaryKey = PrimaryKey(id)

    init {
        uniqueIndex("links_code_key", code)
        index(
            customIndexName = "idx_links_active_by_user_created_at",
            isUnique = false,
            userId,
            createdAt,
            filterCondition = { deletedAt.isNull() },
        )
    }
}
