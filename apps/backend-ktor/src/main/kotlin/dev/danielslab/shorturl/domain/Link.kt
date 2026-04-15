package dev.danielslab.shorturl.domain

import kotlin.time.Clock
import kotlin.time.Instant

data class Link(
    val id: Int,
    val userId: Int,
    val code: String,
    val originalUrl: String,
    val clicks: Int,
    var note: String? = null,
    val createdAt: Instant = Clock.System.now(),
    val deletedAt: Instant? = null,
)
