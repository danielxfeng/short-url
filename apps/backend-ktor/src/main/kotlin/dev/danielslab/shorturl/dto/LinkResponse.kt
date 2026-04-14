package dev.danielslab.shorturl.dto

import dev.danielslab.shorturl.domain.Link

data class LinkResponse(
    val id: Int,
    val userId: Int,
    val code: String,
    val originalUrl: String,
    val clicks: Int,
    val note: String? = null,
    val createdAt: String,
    val isDeleted: Boolean,
) {
    companion object {
        fun fromDomain(link: Link): LinkResponse =
            LinkResponse(
                link.id,
                link.userId,
                link.code,
                link.originalUrl,
                link.clicks,
                link.note,
                link.createdAt.toString(),
                link.deletedAt != null,
            )
    }
}
