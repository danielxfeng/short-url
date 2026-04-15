package dev.danielslab.shorturl.dto.oauth

import dev.danielslab.shorturl.domain.UserProvider
import dev.danielslab.shorturl.repository.core.UserUpsertInput
import kotlinx.serialization.Serializable

interface OAuthUserProfile {
    fun toUserUpsertInput(): UserUpsertInput
}

@Serializable
data class GoogleUserInfo(
    val sub: String,
    val name: String? = null,
    val picture: String? = null,
) : OAuthUserProfile {
    override fun toUserUpsertInput(): UserUpsertInput =
        UserUpsertInput(
            provider = UserProvider.GOOGLE,
            providerId = sub,
            displayName = name,
            profilePicture = picture,
        )
}

@Serializable
data class GithubUserInfo(
    val id: Long,
    val name: String? = null,
    @kotlinx.serialization.SerialName("avatar_url")
    val avatarUrl: String? = null,
) : OAuthUserProfile {
    override fun toUserUpsertInput(): UserUpsertInput =
        UserUpsertInput(
            provider = UserProvider.GITHUB,
            providerId = id.toString(),
            displayName = name,
            profilePicture = avatarUrl,
        )
}
