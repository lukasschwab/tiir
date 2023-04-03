# tiir

A flexible successor to [tir](https://github.com/lukasschwab/tir), but not a drop-in replacement.

## Goals

+ Support multiple backends ("stores").
    + Memory, for testing.
    + Local file.
    + Remote file (i.e. backed up to git).[^bak]
+ Support multiple editor interfaces.
    + Updated simple CLI, like the current interface.[^cb]
        + Improved text editing experience; go back to a field you've already submitted.
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
+ I'm leaning hard on interfaces. Conveniently, Jimmy Koppel posted a bit on their importance: https://www.pathsensitive.com/2023/03/modules-matter-most-for-masses.html
+ Question: am I over-modularizing? I should look at the graph and see if I have modules with single dependents.
+ Text is its own module to avoid cyclical dependencies between pkg/store and pkg/tir.
    + Do I move this to a proto definition eventually? If yes, I can maybe move the various helpers from pkg/text into pkg/tir.
+ Using embed for templates to resolve paths.
    + Google reveals tis is very normal, but I realized it on my own because of [that one quite repo.](https://github.com/eliben/go-quines/blob/main/quine-source-embed.go)
    + The issue: running tests from inside pkg/render worked, but running the app from root did not.
+ File as database is insipired by Tailscale... but I could probably optimize it for long-running server use.

One nice thing about interfaces: defining them early lets you implement things out-of-order. You can write pkg/tir *assuming* something will implement pkg/store. Don't get it quite right? Tweak it later.

To get current tir behavior... I can use this locally and just throw my .tir.json file someplace it'll be backed up by some other daemon (i.e. ~/Documents)! But I have to figure out my publish flow if I want a public site.

For a read-only public site, can either run the fly.io server with a flag *or* I can build static assets and deploy them to e.g. GitHub Pages. Or I have to figure out fly.io authentication.

## MVP

+ "Hosted" tir; I can post things to tir from a work computer.
+ That means figuring out auth.

## More things to learn

+ `GORM` with some SQL store, e.g. Planetscale.
+ Running a backup daemon.
+ Typescript/React frontend. Can go deep, deep down the rabbit-hole and build an OAuth app.
