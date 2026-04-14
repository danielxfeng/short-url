package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.domain.UserProvider
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class UserResponseDto(
    val id: Int,
    val provider: UserProvider,
    @SerialName("provider_id")
    val providerId: String,
    @SerialName("display_name")
    val displayName: String? = null,
    @SerialName("profile_picture")
    val profilePicture: String? = null,
)
