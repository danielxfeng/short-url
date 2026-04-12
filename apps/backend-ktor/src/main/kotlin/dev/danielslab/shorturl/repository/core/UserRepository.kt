package dev.danielslab.shorturl.repository.core

import dev.danielslab.shorturl.domain.User
import dev.danielslab.shorturl.domain.UserUpsertInput

interface UserRepository {
    suspend fun getUserById(id: Int): User?
    suspend fun deleteUserById(id: Int)
    suspend fun upsertUser(userInput: UserUpsertInput)
}
