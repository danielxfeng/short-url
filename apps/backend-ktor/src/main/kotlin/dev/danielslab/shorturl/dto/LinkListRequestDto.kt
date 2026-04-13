package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.repository.core.LinkListInput

data class LinkListRequestDto(
    val cursor: String? = null,
    val limit: Int? = null,
) {
    fun toRepoParam(config: Config, userId: Int): LinkListInput = LinkListInput(
        userId = userId,
        cursor = cursor,
        limit = when {
            limit == null -> config.linkDefaultPageSize
            limit < config.linkDefaultPageSize -> config.linkDefaultPageSize
            limit > config.linkMaxPageSize -> config.linkMaxPageSize
            else -> limit
        },
    )
}
