# CRUD

company crud test task

## Prerequisites

Docker - https://docs.docker.com/engine/install

## Architecture

The project follows [Package Oriented Design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html) (aka Domain Oriented Design). An alternative layout with horizontal packages (mongo, service, handler, etc.) could also be used.

- [cmd](./cmd) - executables
- [company](./company) - public project files
- [internal](./internal) - internal project files
- [migrations](./migrations) - mongoDB migrations
- [mongo](./mongo) - mongoDB init scripts
- [auth](./auth) - JWT auth files

## Testing and Development

To run all tests run `make test`. To run linters run `make lint`.

To build the project execute `make build`.

To run the project locally execute `make start`.

## Migrations
Create new migration: `./mongo/migrate.sh create -ext json -seq -dir path/to/migrations/ migration_name`.

Run migrations: `make migrate`
