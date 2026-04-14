package dev.danielslab.shorturl.dto

import kotlinx.serialization.Serializable

@Serializable
data class LinkCreateRequestDto(
    var code: String? = null,
    val originalUrl: String,
    val note: String? = null,
)
