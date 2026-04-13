package dev.danielslab.shorturl.db.schema

import dev.danielslab.shorturl.db.exposed.PgEnum
import dev.danielslab.shorturl.domain.UserProvider
import org.jetbrains.exposed.sql.Column
import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.javatime.CurrentTimestampWithTimeZone
import org.jetbrains.exposed.sql.javatime.timestampWithTimeZone

object UsersTable : Table("users") {
    val id = integer("id").autoIncrement()
    val provider : Column<UserProvider> = customEnumeration(
        name = "provider",
        sql = "provider_enum",
        fromDb = { value -> UserProvider.valueOf(value as String ) },
        toDb = { PgEnum("provider_enum", it) },
    )
    val providerId = text("provider_id")
    val displayName = text("display_name").nullable()
    val profilePic = text("profile_pic").nullable()
    val createdAt = timestampWithTimeZone("created_at").defaultExpression(CurrentTimestampWithTimeZone)

    override val primaryKey = PrimaryKey(id)

    init {
        uniqueIndex("provider", provider, providerId)
    }
}
