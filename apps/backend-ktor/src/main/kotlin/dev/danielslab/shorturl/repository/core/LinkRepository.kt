package dev.danielslab.shorturl.repository.core

import dev.danielslab.shorturl.domain.Link

data class LinkCreateInput(
    val userId: Int,
    val code: String,
    val originalUrl: String,
    val note: String? = null,
)

data class LinkDeleteInput(
    val userId: Int,
    val code: String,
)

data class LinkListInput(
    val userId: Int,
    val cursor: String? = null,
    val limit: Int,
)

interface LinkRepository {
    suspend fun createLink(request: LinkCreateInput): Link
    suspend fun getLinkByCode(code: String): Link?
    suspend fun getLinksByUserId(request: LinkListInput): List<Link>
    suspend fun softDeleteLinkById(request: LinkDeleteInput)
    suspend fun restoreLinkById(request: LinkDeleteInput)
    suspend fun permanentlyDeleteLinkById(request: LinkDeleteInput)
    suspend fun setLinkClicked(code: String): Int
}