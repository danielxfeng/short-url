package dev.danielslab.shorturl.config

import io.github.cdimascio.dotenv.Dotenv
import java.net.URI

data class Config (
    val port : Int,
    val cors : String,
    val backendPublicUrl : String,
    val dbUrl : String,
    val testDbUrl : String,
    val jwtSecret : String,
    val jwtExpiry : Int,
    val notFoundPage: String,
    val googleClientId: String,
    val googleClientSecret : String,
    val githubClientId: String,
    val githubClientSecret : String,
    val frontendRedirectUrl : String,
) {
    val corsScheme: String
        get() = corsUri.scheme ?: error("CORS origin must include a scheme: $cors")

    val corsHost: String
        get() = corsUri.authority ?: error("CORS origin must include a host: $cors")

    companion object {
        private val dotenv: Dotenv by lazy {
            Dotenv.configure()
                .ignoreIfMissing()
                .load()
        }

        fun fromEnv() : Config = Config(
            port = envInt("PORT", 8081),
            cors = envString("CORS", "http://localhost:5173"),
            backendPublicUrl = envString("BACKEND_PUBLIC_URL", "http://localhost:8080"),
            dbUrl = envString("DB_URL"),
            testDbUrl = envString("TEST_DB_URL"),
            jwtSecret = envString("JWT_SECRET"),
            jwtExpiry = envInt("JWT_EXPIRY", 24 * 7),
            notFoundPage = envString("NOT_FOUND_PAGE", "http://localhost:5173/not-found"),
            googleClientId = envString("GOOGLE_CLIENT_ID"),
            googleClientSecret = envString("GOOGLE_CLIENT_SECRET"),
            githubClientId = envString("GITHUB_CLIENT_ID"),
            githubClientSecret = envString("GITHUB_CLIENT_SECRET"),
            frontendRedirectUrl = envString("FRONTEND_REDIRECT_URL", "http://localhost:5173/auth/callback"),
        )

        private fun envString(key: String, default: String? = null): String =
            System.getenv(key) ?: dotenv[key] ?: default ?: error("Environment variable $key is missing")

        private fun envInt(key: String, default: Int? = null) : Int =
            envString(key, default?.toString()).toIntOrNull()
                ?: error("Environment variable $key is missing")
    }

    private val corsUri: URI
        get() = runCatching { URI(cors) }
            .getOrElse { error("Invalid CORS origin: $cors") }
}
