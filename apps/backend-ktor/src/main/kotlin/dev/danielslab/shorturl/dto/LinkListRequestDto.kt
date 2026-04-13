package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.repository.core.LinkListInput

data class LinkListRequestDto(
    val cursor: String? = null,
    val limit: Int? = null,
) {
    fun toRepoParam(
        userId: Int,
        defaultPageSize: Int,
        maxPageSize: Int,
    ): LinkListInput = LinkListInput(
        userId = userId,
        cursor = cursor,
        limit = when {
            limit == null -> defaultPageSize
            limit < defaultPageSize -> defaultPageSize
            limit > maxPageSize -> maxPageSize
            else -> limit
        },
    )
}
