Prudence
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/prudence.svg)](https://github.com/tliron/prudence/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/prudence.svg)](https://pkg.go.dev/github.com/tliron/prudence)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/prudence)](https://goreportcard.com/report/github.com/tliron/prudence)

A lightweight web framework built for scale, featuring baked-in
[RESTful](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) server-side and
client-side caching. Suitable for frontend user interfaces and backend APIs.

Prudence is distributed as a single, compact, customizable executable file with no external
dependencies.

[![Download](assets/media/download.png "Download")](https://github.com/tliron/prudence/releases)

Web frameworks are a dime a dozen! So what makes this one worthy of your consideration? Well, here's the
pitch: it distills years of professional experience with REST caching, boils down lessons learned (the hard
way) from well-intentioned but over-engineered experiments, finally landing in a from-scratch codebase that
is laser-focused on exactly what needs to be in a framework and nothing else. Prudence is eminently lightweight
and yet eminently expressive. It intends to be the last stop in your search for the right framework. Welcome
home.


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


Feature Highlights
------------------

* Triple-phase representation generation allows for composable, fine-grained, associative control
  over server-side and client-side caching. Reap the full benefits of idempotency in RESTful network
  architectures.
* Prudence is written in Go for reliability and performance but allows you to use straightforward
  JavaScript for assembling and configuring your application. (JavaScript is *not required*.
  [Here's](https://github.com/tliron/prudence/tree/main/examples/go) an example in pure Go.)
* Additionally supports [JavaScript Templates (JST)](jst/README.md) for generating HTML pages by
  combining design with code scriptlets. And there's sugar.
* Pluggable server-side cache backends. Store your generated representations in fast distributed
  stores such as [Memcached](https://memcached.org/), [Redis](https://redis.io/), etc.
* [Extensible](platform/README.md) via the [xprudence tool](xprudence/README.md), which allows you
  to create custom builds of Prudence bundled with the plugins and APIs required by your applications.
  Even when extended in this way, Prudence is still distributed as a single, compact executable file.
* Prudence is fun. Through rigorous benchmarks conducted in our good-times laboratory we found
  Prudence to be a zillion times more fun than competing products from leading brand names.
