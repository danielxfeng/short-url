package dev.danielslab.shorturl.dto

import kotlinx.serialization.Serializable

@Serializable
data class LinkResponse(
    val id: Int,
    val userId: Int,
    val code: String,
    val originalUrl: String,
    val clicks: Int,
    val note: String? = null,
    val createdAt: String,
    val isDeleted: Boolean,
)
