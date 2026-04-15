package dev.danielslab.shorturl.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LinkCreateRequestDto(
    var code: String? = null,
    @SerialName("original_url")
    val originalUrl: String,
    val note: String? = null,
)
