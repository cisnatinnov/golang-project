# Backend Engineering Interview Assignment (Golang)

## Requirements

To run this project you need to have the following installed:

1. [Go](https://golang.org/doc/install) version 1.21
2. [GNU Make](https://www.gnu.org/software/make/)
3. [oapi-codegen](https://github.com/deepmap/oapi-codegen)

    Install the latest version with:
    ```
    go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
    ```
4. [mock](https://github.com/uber-go/mock)

    Install the latest version with:
    ```
    go install go.uber.org/mock/mockgen@latest
    ```

5. [Docker](https://docs.docker.com/get-docker/) version 20
   
   We will use this for testing your API.

6. [Docker Compose](https://docs.docker.com/compose/install/) version 1.29

7. [Node](https://nodejs.org/en) v20

   We will use this for testing your API

8. [NPM](https://www.npmjs.com/) v10

    We will use this for testing your API.

## Project Structure

The project follows a standard Go hexagonal-style layout:

```text
api.yml             ← OpenAPI 3.0 spec (Source of truth)
generated/          ← Code generated from api.yml (DO NOT EDIT)
cmd/main.go         ← Entry point for the application
handler/            ← HTTP request handlers (Echo)
  server.go         ← Server configuration and DI
  endpoints.go      ← Implementation of API endpoints
repository/         ← Data access layer (PostgreSQL)
  interfaces.go     ← Repository interface definitions
  implementations.go← Database-specific implementations
  types.go          ← Repo layer data models
database.sql        ← Database schema definitions
tests/              ← Integration tests
Makefile            ← Build, test, and code generation targets
```

## Development Workflow

To add a new feature, follow this standard workflow:

1.  **Define the API**: Edit `api.yml` with the new endpoint/schema.
2.  **Generate Code**: Run `make generate` to update `generated/`.
3.  **Define Repository**: Add the new method to `repository/interfaces.go`.
4.  **Implement Repository**: Add the logic in `repository/implementations.go`.
5.  **Generate Mocks**: Run `make generate_mocks` for unit testing.
6.  **Implement Handler**: Implement the handler in `handler/endpoints.go`.
7.  **Test**: Add unit tests in `handler/endpoints_test.go` and integration tests in `tests/api_test.go`.

## Initiate The Project

To start working, execute

```
make init
```

## Running

You should be able to run using the script `run.sh`:

```bash
./run.sh
```

You may see some errors since you have not created the API yet.

However for testing, you can use Docker run the project, run the following command:

```
docker-compose up --build
```

You should be able to access the API at http://localhost:8080

If you change `database.sql` file, you need to reinitate the database by running:

```
docker compose down --volumes
```

## Testing

To run test, run the following command:

```
make test
```

---

For more detailed technical conventions and architecture rules, please refer to [AGENTS.md](AGENTS.md).
