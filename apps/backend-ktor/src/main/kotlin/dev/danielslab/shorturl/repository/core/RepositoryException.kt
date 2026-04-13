package dev.danielslab.shorturl.repository.core

class NotFoundException(message: String) : NoSuchElementException(message)

class DuplicatedException(message: String) : IllegalStateException(message)
