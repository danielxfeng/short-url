package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.domain.UserUpsertInput
import kotlinx.serialization.Serializable

@Serializable
data class UserUpsertRequestDto(
    val provider: UserProvider,
    val providerId: String,
    val displayName: String? = null,
    val profilePicture: String? = null,
) {
    fun toDomain(): UserUpsertInput =
        UserUpsertInput(provider, providerId, displayName, profilePicture)
}
