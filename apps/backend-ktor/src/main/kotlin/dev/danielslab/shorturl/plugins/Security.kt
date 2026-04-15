package dev.danielslab.shorturl.plugins

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import dev.danielslab.shorturl.config.Config
import io.ktor.client.HttpClient
import io.ktor.client.engine.apache5.Apache5
import io.ktor.http.HttpMethod
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.auth.Authentication
import io.ktor.server.auth.OAuthServerSettings
import io.ktor.server.auth.jwt.JWTPrincipal
import io.ktor.server.auth.jwt.jwt
import io.ktor.server.auth.oauth

fun Application.configureSecurity(config: Config) {
    install(Authentication) {
        jwt("auth-jwt") {
            realm = config.jwtRealm
            verifier(
                JWT
                    .require(Algorithm.HMAC256(config.jwtSecret))
                    .withAudience(config.jwtAudience)
                    .withIssuer(config.jwtDomain)
                    .build(),
            )
            validate { credential ->
                if (credential.payload.audience.contains(config.jwtAudience)) JWTPrincipal(credential.payload) else null
            }
        }
        oauth("auth-oauth-google") {
            urlProvider = { "${config.backendPublicUrl}/api/v1/user/auth/google/callback" }
            providerLookup = {
                OAuthServerSettings.OAuth2ServerSettings(
                    name = "google",
                    authorizeUrl = config.googleAuthorizeUrl,
                    accessTokenUrl = config.googleAccessTokenUrl,
                    requestMethod = HttpMethod.Post,
                    clientId = config.googleClientId,
                    clientSecret = config.googleClientSecret,
                    defaultScopes = config.googleScopes.split(","),
                )
            }
            client = HttpClient(Apache5)
        }
        oauth("auth-oauth-github") {
            urlProvider = { "${config.backendPublicUrl}/api/v1/user/auth/github/callback" }
            providerLookup = {
                OAuthServerSettings.OAuth2ServerSettings(
                    name = "github",
                    authorizeUrl = config.githubAuthorizeUrl,
                    accessTokenUrl = config.githubAccessTokenUrl,
                    requestMethod = HttpMethod.Post,
                    clientId = config.githubClientId,
                    clientSecret = config.githubClientSecret,
                    defaultScopes = config.githubScopes.split(","),
                )
            }
            client = HttpClient(Apache5)
        }
    }
}
