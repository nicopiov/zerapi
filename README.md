# Zerapi

Instant APIs. Zero setup.

Zerapi is a small CLI tool that turns JSON files into temporary local REST APIs for frontend development, prototyping, demos, and tests.

```sh
zerapi serve examples/users.json
```

## What It Does

Zerapi loads a JSON file, infers API resources, stores the data in memory, and exposes CRUD routes locally.

Supported JSON shapes:

```json
[
  { "id": 1, "name": "Ada Lovelace" },
  { "id": 2, "name": "Grace Hopper" }
]
```

Top-level arrays infer the resource name from the file name. For example, `users.json` becomes `/users`.

```json
{
  "users": [{ "id": 1, "name": "Ada Lovelace" }],
  "posts": [{ "id": 1, "title": "Hello" }]
}
```

Top-level objects expose each array property as a resource.

## Install

Install the latest release:

```sh
go install github.com/nicopiov/zerapi@latest
```

Or install a specific version:

```sh
go install github.com/nicopiov/zerapi@v0.2.0
```

Then verify the command is available:

```sh
zerapi version
```

Version aliases are also available:

```sh
zerapi --version
zerapi -v
```

## Usage

Start an API from the example data:

```sh
zerapi serve examples/users.json
```

Use a different port:

```sh
zerapi serve --port 9090 examples/users.json
```

Use a different host:

```sh
zerapi serve --host 127.0.0.1 examples/users.json
```

Start in readonly mode:

```sh
zerapi serve --readonly examples/users.json
```

Reload the file when it changes:

```sh
zerapi serve --watch examples/users.json
```

Show serve help:

```sh
zerapi serve --help
```

## REST Routes

For a `users` resource, Zerapi exposes:

```text
GET    /users
GET    /users/{id}
POST   /users
PUT    /users/{id}
PATCH  /users/{id}
DELETE /users/{id}
```

Example requests:

```sh
curl http://localhost:8080/users
```

```sh
curl http://localhost:8080/users/1
```

```sh
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Linus Torvalds","email":"linus@example.com"}'
```

```sh
curl -X PUT http://localhost:8080/users/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","email":"ada@history.example"}'
```

```sh
curl -X PATCH http://localhost:8080/users/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Byron"}'
```

```sh
curl -X DELETE http://localhost:8080/users/1
```

## Filtering

Collection routes support exact-match query filters:

```sh
curl "http://localhost:8080/users?email=ada@example.com"
```

Multiple filters are combined with AND behavior:

```sh
curl "http://localhost:8080/users?id=1&email=ada@example.com"
```

For values with spaces or other special characters, URL-encode the value:

```sh
curl "http://localhost:8080/users?name=Ada%20Lovelace"
```

Or let `curl` encode it:

```sh
curl --get "http://localhost:8080/users" \
  --data-urlencode "name=Ada Lovelace"
```

## Pagination

Use `_limit` to cap collection results:

```sh
curl "http://localhost:8080/users?_limit=1"
```

Use `_page` with `_limit` to page through results:

```sh
curl "http://localhost:8080/users?_page=2&_limit=1"
```

Collection responses include `X-Total-Count`, which reports the number of matching records before pagination:

```sh
curl -i "http://localhost:8080/users?_limit=1"
```

## Sorting

Use `_sort` to order collection results by a field:

```sh
curl "http://localhost:8080/users?_sort=name"
```

Prefix the field with `-` for descending order:

```sh
curl "http://localhost:8080/users?_sort=-name"
```

## Readonly Mode

Readonly mode keeps loaded data stable by allowing reads and blocking writes.

```sh
zerapi serve --readonly examples/users.json
```

Allowed:

```text
GET /users
GET /users/1
```

Blocked with `403 Forbidden`:

```text
POST   /users
PUT    /users/1
PATCH  /users/1
DELETE /users/1
```

## Watch Mode

Watch mode reloads the source file while the server keeps running.

```sh
zerapi serve --watch examples/users.json
```

When the JSON file changes, Zerapi reloads the in-memory API data. If the new file content is invalid, Zerapi logs a warning and keeps serving the last valid data.

## Development

Run tests:

```sh
go test ./...
```

Install from a local checkout:

```sh
go install .
```

Run without installing:

```sh
go run . version
go run . serve examples/users.json
```

Build a local binary:

```sh
go build -o bin/zerapi .
./bin/zerapi version
```

## Troubleshooting

### Command Not Found

If `zerapi` is not found after `go install`, make sure Go's bin directory is in your `PATH`.

For zsh:

```sh
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Then try:

```sh
zerapi version
```
