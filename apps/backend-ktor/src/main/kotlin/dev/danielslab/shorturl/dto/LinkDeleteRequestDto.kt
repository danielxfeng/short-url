package dev.danielslab.shorturl.dto

import kotlinx.serialization.Serializable

@Serializable
data class LinkDeleteRequestDto(
    val code: String,
)
