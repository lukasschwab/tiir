# tiir

A flexible successor to [tir](https://github.com/lukasschwab/tir), but not a drop-in replacement.

## Goals

+ Support multiple backends ("stores").
    + Memory, for testing.
    + Local file.
    + Remote file (i.e. backed up to git).[^bak]
+ Support multiple editor interfaces.
    + Updated simple CLI, like the current interface.[^cb]
    + `vim` editor for JSON, like I use in [id3ed](https://github.com/lukasschwab/id3ed).
    + REST API service.[^tailscale]
+ Support multiple representations; most of these are already rendered by existing tir.
    + HTML table.
    + Atom feed, JSON feed.
    + Plaintext

[^bak]: Backups may really be a useful concept here, but introduce the issue of reconciliation.

[^cb]: Maybe use [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) to spruce it up. This is good if I want to support more interaction (e.g. *browsing* existing texts before updating them), but may be overkill if I just need simple inputs.

[^tailscale]: Might be fun to expose this via Tailscale. Might be overkill, especially if "remote file" works okay.

This all probably requires a local config pattern.

It will also require new setup instructions, perhaps for two modes (HTTP server)

Finally, it's an opportunity to learn new corners of Go.

+ `text/template` and associated APIs.
+ [Fiber](https://gofiber.io/) may provide more pluggable HTTP patterns, though I'm fond of vanilla HTTP.
+ CLI utilities like the [Charm](https://github.com/charmbracelet) family.
+ [Fly.io](https://fly.io/docs/languages-and-frameworks/golang/) and/or Tailscale for hosting.

## Interesting decisions

+ I'm gravitating towards single-method interfaces; [thanks Eli Bendersky.](https://eli.thegreenplace.net/2023/the-power-of-single-method-interfaces-in-go/)
+ I'm moving most of the `text.Text` manipulation into `text`, even if it only has a single caller. It's nice to put the pluggable types in one place. You shouldn't need `pkg/edit` to override its built-in editors.
    + Ultimately, the cmd dependencies on `pkg/edit` may get moved to `tir`... or wherever we do the config processing.