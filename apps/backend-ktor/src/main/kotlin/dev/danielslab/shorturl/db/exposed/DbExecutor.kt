package dev.danielslab.shorturl.db.exposed

import kotlinx.coroutines.Dispatchers
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.transactions.experimental.newSuspendedTransaction

class DbExecutor(
    private val database: Database,
) {
    suspend fun <T> query(block: suspend () -> T): T =
        newSuspendedTransaction(Dispatchers.IO, db = database) {
            block()
        }
}
