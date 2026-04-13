package dev.danielslab.shorturl

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.plugins.configureRateLimiting
import dev.danielslab.shorturl.plugins.configureRequestValidation
import dev.danielslab.shorturl.plugins.configureStatusPages
import dev.danielslab.shorturl.plugins.configureHTTP
import dev.danielslab.shorturl.plugins.configureMonitoring
import dev.danielslab.shorturl.routes.configureRouting
import dev.danielslab.shorturl.plugins.configureSecurity
import dev.danielslab.shorturl.plugins.configureSerialization
import io.ktor.server.application.*
import io.ktor.server.netty.EngineMain

fun main(args: Array<String>) {
    EngineMain.main(args)
}

fun Application.module() {
    val config = Config.fromEnv(this)

    configureHTTP(
        corsHost = config.corsHost,
        corsScheme = config.corsScheme,
    )
    configureSecurity(
        jwtSecret = config.jwtSecret,
        backendPublicUrl = config.backendPublicUrl,
        googleClientId = config.googleClientId,
        googleClientSecret = config.googleClientSecret,
    )
    configureMonitoring()
    configureSerialization()
    configureRequestValidation()
    configureStatusPages()
    configureRateLimiting()
    configureRouting()
}
