XPrudence: The Prudence Customizer
==================================

The downloadable distribution of Prudence contains a ready-to-use `prudence` executable
that isbundled with all the basic types, APIs, cache backends, renderers, and JST tags.
However, if you want to [extend](../platform/README.md) Prudence then you'll need to make
a custom build of `prudence` with your plugins.

That's what `xprudence`, which is also included in the distribution, is for.

It's a rather simple program! It requires you to have
[Go installed](https://golang.org/doc/install), but otherwise has no dependencies.
All it does is create custom builds of `prudence`.

### Background

This necessity might seem strange to you if you're familiar with web frameforks or indeed
other platforms built on interpreted languages such as Python, Ruby, or Node.js's JavaScript.
In those cases your code would simply "pick up" any other code if it's in the right path
and you could use some kind of package manager (pip, gem, npm, etc.) to add libraries. Even
Java, a compiled language, picks up JAR files on-the-fly and can install them using Maven.

But Go is different, in a good way. Instead of a mess of dependencies scattered all
over your operating system, everything a Go program normally needs is in one executable.
That's just one single file that you need to distribute.

In recognizing this advantage, `xprudence` is inspired by
[xcaddy](https://github.com/caddyserver/xcaddy).

(Note that Go does support building and loading dynamically-linked libraries as separate
files. It's a feature intended mainly for interaction with C and other programming
languages, not for assembling Go programs.)

### Example

Here's how you would build Prudence with the
[plugin example](https://github.com/tliron/prudence/tree/main/examples/plugin):

    xprudence build -v --directory=examples/plugin

The first time you build it may take some time, but note that most of this time is spent
downloading (and caching) the Go source code of the dependencies. The Go compiler/linker
itself is very fast and subsequent builds will use the cached downloads.

A common flag to use with `xprudence build` is `--directory`, which you can use several
times to specify plugin directories that you want to add to your build. These directories
are [Go modules](https://golang.org/ref/mod) and thus must have a `go.mod` file. The
[extension guide](../platform/README.md) goes into more detail about their structure.

Or, you can use the `--module` flag to specify plugins as Go module names, which is useful
if you keep your module code in a git repository. As is usual with Go modules, you can
append a specific version or tag after a `@`.

Another useful flag is `--replace`, which lets you insert a
[replace directive](https://golang.org/ref/mod#go-mod-file-replace).

To get help on the additional build flags:

    xprudence build -h
