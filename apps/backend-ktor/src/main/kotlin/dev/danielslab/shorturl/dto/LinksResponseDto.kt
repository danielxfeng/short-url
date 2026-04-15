package dev.danielslab.shorturl.dto

import kotlinx.serialization.Serializable

@Serializable
data class LinksResponseDto(
    val links: List<LinkResponseDto>,
    val hasMore: Boolean,
    val cursor: Int?,
)
