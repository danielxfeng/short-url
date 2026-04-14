package dev.danielslab.shorturl.routes

import dev.danielslab.shorturl.config.Config
import dev.danielslab.shorturl.repository.core.LinkRepository
import dev.danielslab.shorturl.repository.core.UserRepository
import io.ktor.server.application.Application
import io.ktor.server.routing.route
import io.ktor.server.routing.routing

fun Application.configureRouting(
    config: Config,
    userRepository: UserRepository,
    linkRepository: LinkRepository,
) {
    routing {
        route("/api/v1") {
            route("/health") {
                healthRoutes()
            }
            route("/user") {
                userRoutes(config, userRepository)
            }
            route("/short-urls") {
                linkRoutes(config, linkRepository)
            }
        }
    }
}
