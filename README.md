# tiir ![example branch parameter](https://github.com/lukasschwab/tiir/actions/workflows/go.yml/badge.svg?branch=main)


A flexible successor to [tir](https://github.com/lukasschwab/tir).

## Setup

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

Optionally, see [Fly.io's documentation](https://fly.io/docs/languages-and-frameworks/golang/) for deploying the server with `flyctl launch`.

If you expose your server to the internet, you should secure endpoints modifying your data with an API key. Generate a secret, then set it in your Fly app's environment:

```console
$ flyctl secrets set TIR_API_SECRET=YOUR_SECRET_HERE
```

## Configuration

<!-- TODO: describe how the user can specify these values. Mostly command line arguments -->

`tir` looks for a configuration file at `/etc/tir/.tir.config` and `$HOME/.tir.config`.

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

### Server

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
    }
}
