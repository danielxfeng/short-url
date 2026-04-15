# shorturl

This project was created using the [Ktor Project Generator](https://start.ktor.io).

Here are some useful links to get you started:

- [Ktor Documentation](https://ktor.io/docs/home.html)
- [Ktor GitHub page](https://github.com/ktorio/ktor)
- The [Ktor Slack chat](https://app.slack.com/client/T09229ZC6/C0A974TJ9). You'll need
  to [request an invite](https://surveys.jetbrains.com/s3/kotlin-slack-sign-up) to join.

## Features

Here's a list of features included in this project:

| Name                                                                   | Description                                                                        |
| ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| [Compression](https://start.ktor.io/p/compression)                     | Compresses responses using encoding algorithms like GZIP                           |
| [CORS](https://start.ktor.io/p/cors)                                   | Enables Cross-Origin Resource Sharing (CORS)                                       |
| [Routing](https://start.ktor.io/p/routing)                             | Provides a structured routing DSL                                                  |
| [OpenAPI](https://start.ktor.io/p/openapi)                             | Serves OpenAPI documentation                                                       |
| [Authentication](https://start.ktor.io/p/auth)                         | Provides extension point for handling the Authorization header                     |
| [Authentication JWT](https://start.ktor.io/p/auth-jwt)                 | Handles JSON Web Token (JWT) bearer authentication scheme                          |
| [Authentication OAuth](https://start.ktor.io/p/auth-oauth)             | Handles OAuth Bearer authentication scheme                                         |
| [Request Validation](https://start.ktor.io/p/request-validation)       | Adds validation for incoming requests                                              |
| [Call Logging](https://start.ktor.io/p/call-logging)                   | Logs client requests                                                               |
| [Call ID](https://start.ktor.io/p/callid)                              | Allows to identify a request/call.                                                 |
| [Content Negotiation](https://start.ktor.io/p/content-negotiation)     | Provides automatic content conversion according to Content-Type and Accept headers |
| [kotlinx.serialization](https://start.ktor.io/p/kotlinx-serialization) | Handles JSON serialization using kotlinx.serialization library                     |
| [Exposed](https://start.ktor.io/p/exposed)                             | Adds Exposed database to your application                                          |
| [Postgres](https://start.ktor.io/p/postgres)                           | Adds Postgres database to your application                                         |
| [Rate Limiting](https://start.ktor.io/p/ktor-server-rate-limiting)     | Manage request rate limiting as you see fit                                        |
| [HSTS](https://start.ktor.io/p/hsts)                                   | Enables HTTP Strict Transport Security (HSTS)                                      |
| [Status Pages](https://start.ktor.io/p/status-pages)                   | Provides exception handling for routes                                             |

## Building & Running

To build or run the project, use one of the following tasks:

| Task                                    | Description                                                          |
| --------------------------------------- | -------------------------------------------------------------------- |
| `./gradlew test`                        | Run the tests                                                        |
| `./gradlew build`                       | Build everything                                                     |
| `./gradlew buildFatJar`                 | Build an executable JAR of the server with all dependencies included |
| `./gradlew buildImage`                  | Build the docker image to use with the fat JAR                       |
| `./gradlew publishImageToLocalRegistry` | Publish the docker image locally                                     |
| `./gradlew run`                         | Run the server                                                       |
| `./gradlew runDocker`                   | Run using the local docker image                                     |

## Monorepo Commands

This app is also exposed through the workspace `pnpm` setup so CI and local commands can use the same entrypoints as the other apps:

| Task                                              | Description                            |
| ------------------------------------------------- | -------------------------------------- |
| `pnpm --filter backend-ktor test`                 | Run tests                              |
| `pnpm --filter backend-ktor lint`                 | Run ktlint checks                      |
| `pnpm --filter backend-ktor format`               | Auto-format Kotlin sources with ktlint |
| `pnpm --filter backend-ktor build`                | Build the application                  |
| `pnpm --filter backend-ktor start`                | Run the Ktor server locally            |
| `pnpm --filter backend-ktor jar`                  | Build the executable fat JAR           |
| `pnpm --filter backend-ktor docker:build`         | Build the container image              |
| `pnpm --filter backend-ktor docker:publish:local` | Publish the image to a local registry  |

The repository root Husky pre-commit hook runs `lint-staged`, which now auto-formats staged Kotlin files in `apps/backend-ktor` with `ktlintFormat`.

## CI And Deploy

- Pull requests touching `apps/backend-ktor/**` run `.github/workflows/ci-backend-ktor.yml`.
- The workflow sets up Java 21 and runs `./gradlew test build` directly from `apps/backend-ktor`.
- For deployment artifacts, use `pnpm --filter backend-ktor jar` for a fat JAR or `pnpm --filter backend-ktor docker:build` for a container image.
- Registry publishing is intentionally left provider-specific. Once you choose a registry, the deploy job can extend the CI workflow with Docker login and `docker push` after `buildImage`.

If the server starts successfully, you'll see the following output:

```
2024-12-04 14:32:45.584 [main] INFO  Application - Application started in 0.303 seconds.
2024-12-04 14:32:45.682 [main] INFO  Application - Responding at http://0.0.0.0:8080
```
