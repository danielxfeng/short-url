package dev.danielslab.shorturl.dto

data class LinkListRequestDto(
    val cursor: String? = null,
    val limit: String? = null,
)
