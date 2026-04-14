package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.domain.UserProvider
import kotlinx.serialization.Serializable

@Serializable
data class UserResponseDto(
    val id: Int,
    val provider: UserProvider,
    val providerId: String,
    val displayName: String? = null,
    val profilePicture: String? = null,
) {
    companion object {
        fun fromDomain(user: User): UserResponseDto =
            UserResponseDto(
                user.id,
                user.provider,
                user.providerId,
                user.displayName,
                user.profilePicture,
            )
    }
}
