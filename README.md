# tiir [![Go Reference](https://pkg.go.dev/badge/github.com/lukasschwab/tiir.svg)](https://pkg.go.dev/github.com/lukasschwab/tiir) ![golangci-lint status](https://github.com/lukasschwab/tiir/actions/workflows/go.yml/badge.svg?branch=main)


A flexible successor to [tir](https://github.com/lukasschwab/tir):

> tir – short for "Today I Read" – is a barebones CLI for logging memorable articles.
>
> ![GIF from "One Hundred and One Dalmatians" (1961) of Anita Radcliffe reading a book on a park bench.](https://camo.githubusercontent.com/2d251a5ede6b1cc7ac83876897c7743eadd30202f9343680b116569ee8d5367c/687474703a2f2f6c756b61737363687761622e6769746875622e696f2f696d672f72656164696e672e676966)

## Setup

I recommend [hosting a tir server on the internet](#http-server), e.g. with Fly.io, and configuring your local tir CLI to [store texts through that server](#remote-server).

### CLI

Install the CLI with `go install`:

```console
$ go install ./cmd/tir
```

By default, `tir` is configured to use the rich CLI interface (see [pkg/edit/tea.go](./pkg/edit/tea.go)) and store your data in `$HOME/.tir.json` (see [pkg/store/file.go](pkg/store/file.go)).

To override those defaults, see [Configuration](#configuration).

For CLI documentation, run `tir help`.

### HTTP server

The tir server is an HTTP interface for a store. You can point a store.HTTP at a running server instance to use its store over HTTP.

To run a server locally, run:

```console
$ go run ./cmd/server
```

Optionally, see [Fly.io's documentation](https://fly.io/docs/languages-and-frameworks/golang/) for deploying the server with `flyctl launch`. That process should prompt you to create a volume, which will store (and automatically back up) your tir database.

If you expose your server to the internet, you should secure endpoints modifying your data with an API key. Generate a secret, then set it in your Fly app's environment:

```console
$ flyctl secrets set TIR_API_SECRET=YOUR_SECRET_HERE
```

## Configuration

`tir` looks for a configuration file at `/etc/tir/.tir.config` and `$HOME/.tir.config`. This file configures two independent components:

1. *Store* for persisting and retrieving texts.
2. *Editor* for viewing and modifying texts.

Some example configurations are provided below. Run `tir help` or read [./pkg/config/config.go](./pkg/config/config.go) for more details.

### Local file store

This `.tir.config` file configures tir to use a file store rooted at `/Users/me/tir.json`, to use `vim` to author and edit stored texts:

```json
{
    "store": {
        "type": "file",
        "path": "/Users/me/.tir.json"
    },
    "editor": "vim"
}
```

### libSQL and SQLite3 databases

[libSQL](https://libsql.org/) is an open-source fork of SQLite maintained by [Turso](https://turso.tech/); it retains SQLite's features, adds extensions *not used by this project,* and provides a [`database/sql`-compatible SQLite driver](https://github.com/libsql/libsql-client-go/).

This powers two kinds of stores, comparable to [local files](#local-file-store) and [remote servers](#remote-server) respectively, with the same `connection_string` config input.

#### Local SQLite3 database file

SQLite databases are usually single files on disk.[^file] You can use a local SQLite file *in lieu* of a JSON file ([Local file store](#local-file-store) above).

[^file]: https://www.sqlite.org/fileformat.html

```json
{
    "store": {
        "type": "libsql",
        "connection_string": "file:///path/to/file.db"
    }
}
```

If you supply an initially empty file, tir will initialize the required database table automatically.

#### Turso-hosted `libSQL` database 

```json
{
    "store": {
        "type": "libsql",
        "connection_string": "libsql://[your-database].turso.io?authToken=[your-auth-token]"
    }
}
```

See [Turso's documentation](https://docs.turso.tech/reference/turso-cli#database-client-authentication-tokens) for instructions on how to generate your database auth token.

If you supply a read-only token, tir won't be able to initialize the database or create, update, or delete texts. That offers redundant security on a hosted [HTTP server](#http-server), but it's probably *not* what you want locally.

### `cmd/server` instance

This `.tir.config` file configures tir to talk to a server at `https://tir.fly.dev/` that accepts the API secret `YOUR_API_SECRET`, and to use the rich CLI editor:

```json
{
    "store": {
        "type": "http",
        "base_url": "https://tir.fly.dev/",
        "api_secret": "YOUR_API_SECRET"
    },
    "editor": "tea"
}
```

Alternatively, if you're running the server locally on port 8080:

```json
{
    "store": {
        "type": "http",
        "base_url": "localhost:8080",
        "api_secret": "YOUR_API_SECRET"
    },
    "editor": "tea"
}
```
