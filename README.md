Prudence
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/prudence.svg)](https://github.com/tliron/prudence/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/prudence.svg)](https://pkg.go.dev/github.com/tliron/prudence)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/prudence)](https://goreportcard.com/report/github.com/tliron/prudence)

A lightweight web framework built for scale, featuring baked-in
[RESTful](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) server-side and
client-side caching. Suitable for frontend user interfaces and backend APIs.

Prudence itself is distributed as a single, compact executable file with no external dependencies.

[![Download](assets/media/download.png "Download")](https://github.com/tliron/prudence/releases)


Documentation
-------------

* [Tutorial](TUTORIAL.md)
* [Caching guide](CACHING.md)
* [Examples](examples/README.md)
* [JavaScript API](js/README.md)
* [JavaScript Templates (JST)](jst/README.md)
* [Renderers](render/README.md)
* [Extension guide](platform/README.md)
* [xprudence](xprudence/README.md)
* [Go API](https://pkg.go.dev/github.com/tliron/prudence)
* [FAQ](FAQ.md)


Highlights
----------

* Triple-phase representation generation allows for composable, fine-grained, associative control
  over server-side and client-side caching. Reap the full benefits of idempotency in RESTful network
  architectures.
* Pluggable server-side cache backends. Store your generated representations in fast distributed
  stores such as [Memcached](https://memcached.org/), [Redis](https://redis.io/), etc.
* [Extensible](platform/README.md) via the [xprudence tool](xprudence/README.md), which allows you
  to create custom builds of Prudence bundled with the plugins, modules, and APIs required by your
  applications. Even when extended in this way, Prudence is still distributed as a single, compact
  executable file.
* Uses [fasthttp](https://github.com/valyala/fasthttp) for allocation-free handling of HTTP
  requests and high-performance serving of static files.
* Prudence is written in Go for reliability and performance but allows you to use straightforward
  JavaScript for assembling and configuring your application. (JavaScript is *not required*.
  [Here's](https://github.com/tliron/prudence/tree/main/examples/go) an example in pure Go.)
* Additionally supports extensible [JavaScript Templates (JST)](jst/README.md) for generating HTML
  pages by combining design and programming.
* Prudence is fun. Through rigorous benchmarks conducted in our good-times laboratory we found
  Prudence to be a zillion times more fun than products from leading brand names.
