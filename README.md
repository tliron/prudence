Prudence
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/prudence-go.svg)](https://github.com/tliron/prudence-go/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/prudence-go)](https://goreportcard.com/report/github.com/tliron/prudence-go)

![Prudence](assets/media/prudence.png "Prudence")

A lightweight web framework built for scale, featuring baked-in RESTful client-side and server-side
caching capabilities. Suitable for building frontend user interfaces and backend APIs.

Prudence itself is distributed as a single, compact executable file with no external dependencies.

[![Download](assets/media/download.png "Download")](https://github.com/tliron/prudence-go/releases)

Start with the detailed [quickstart guide](QUICKSTART.md)!


Highlights
----------

* Prudence is written in Go for reliability and performance but allows you to use straightforward
  JavaScript for assembling and configuring your application.
* It additionally supports extensible JavaScript Templates (JST) for generating HTML pages with a
  manageable combination of design and programming.
* Triple-phase representation generation allows for composable, fine-grained control over
  server-side and client-side caching. Reap the full scalability benefits of REST network
  architecture.
* Pluggable server-side cache backends. Store your generated representations in scalable, fast
  stores such as [Memcached](https://memcached.org/), [Redis](https://redis.io/), etc.
* [Extensible](platform/README.md) via the [xprudence tool](xprudence/README.md), which allows you
  to create custom builds of Prudence bundled with the plugins, modules, and APIs required by your
  applications. Even when extended, Prudence is still distributed as a single, compact executable
  file.
* Builds on [fasthttp](https://github.com/valyala/fasthttp) for high-performance handling of HTTP
  requests, including serving static files.
