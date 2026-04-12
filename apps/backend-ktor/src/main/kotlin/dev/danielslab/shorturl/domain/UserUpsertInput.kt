package dev.danielslab.shorturl.domain

data class UserUpsertInput(
    val provider: UserProvider,
    val providerId: String,
    val displayName: String? = null,
    val profilePicture: String? = null,
)
