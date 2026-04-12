package dev.danielslab.shorturl.domain

import kotlin.time.Clock
import kotlin.time.Instant

data class User (
    val id: Int,
    val provider: UserProvider,
    val providerId: String,
    val displayName: String? = null,
    val profilePicture: String? = null,
    val createdAt: Instant = Clock.System.now(),
)