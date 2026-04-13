package dev.danielslab.shorturl.repository.core

import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.domain.UserProvider

data class UserUpsertInput(
    val provider: UserProvider,
    val providerId: String,
    val displayName: String? = null,
    val profilePicture: String? = null,
)

interface UserRepository {
    suspend fun getUserById(id: Int): User?
    suspend fun deleteUserById(id: Int)
    suspend fun upsertUser(userInput: UserUpsertInput)
}
