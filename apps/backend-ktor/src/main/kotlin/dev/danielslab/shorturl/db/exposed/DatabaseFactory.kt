package dev.danielslab.shorturl.db.exposed

import org.jetbrains.exposed.sql.Database

class DatabaseFactory(
    private val driver: String = "org.postgresql.Driver",
) {
    fun connect(url: String): Database =
        Database.connect(
            url = url,
            driver = driver,
        )
}
