package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.repository.core.LinkCreateInput
import kotlinx.serialization.Serializable

@Serializable
data class LinkCreateRequestDto(
    val code: String,
    val originalUrl: String,
    val note: String? = null,
) {
    fun toRepoParam(userId: Int): LinkCreateInput = LinkCreateInput(userId, code, originalUrl, note)
}
