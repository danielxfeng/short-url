package dev.danielslab.shorturl.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LinksResponseDto(
    val links: List<LinkResponseDto>,
    @SerialName("has_more")
    val hasMore: Boolean,
    val cursor: Int?,
)
