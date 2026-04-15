package dev.danielslab.shorturl.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LinkResponseDto(
    val id: Int,
    val code: String,
    @SerialName("original_url")
    val originalUrl: String,
    val clicks: Int,
    val note: String? = null,
    @SerialName("created_at")
    val createdAt: String,
    @SerialName("is_deleted")
    val isDeleted: Boolean,
)
