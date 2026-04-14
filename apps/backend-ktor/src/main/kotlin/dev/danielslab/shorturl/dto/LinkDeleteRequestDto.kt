package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.repository.core.LinkDeleteInput
import kotlinx.serialization.Serializable

@Serializable
data class LinkDeleteRequestDto(
    val code: String,
) {
    fun toRepoParam(userId: Int): LinkDeleteInput = LinkDeleteInput(userId, code)
}
