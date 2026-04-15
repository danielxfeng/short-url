package dev.danielslab.shorturl.db.exposed

import org.postgresql.util.PGobject

class PgEnum<T : Enum<T>>(
    enumTypeName: String,
    enumValue: T?,
) : PGobject() {
    init {
        type = enumTypeName
        value = enumValue?.name
    }
}
