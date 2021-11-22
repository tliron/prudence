Prudence
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/prudence.svg)](https://github.com/tliron/prudence/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/prudence.svg)](https://pkg.go.dev/github.com/tliron/prudence)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/prudence)](https://goreportcard.com/report/github.com/tliron/prudence)

An opinionated lightweight web framework built for scale, featuring integrated
[RESTful](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) server-side and
client-side caching. Write your code in JavaScript, TypeScript, or Go. Suitable for frontend user
interfaces and backend APIs.

Prudence is distributed as a single, compact, customizable executable file with no external dependencies.
And it's fun! Through rigorous benchmarks conducted in our good-times laboratory we found Prudence to be
a zillion times more fun than competing products from leading brand names.

[![Download](assets/media/download.png "Download")](https://github.com/tliron/prudence/releases)


Documentation
-------------

* [Tutorial](TUTORIAL.md)
* [Caching guide](CACHING.md)
* [Examples](examples/README.md)
* [JavaScript/TypeScript API](https://prudence.threecrickets.com/assets/typescript/prudence/docs/)
* [JavaScript Templates (JST)](jst/README.md)
* [Renderers](render/README.md)
* [Extension guide](platform/README.md)
* [xprudence](xprudence/README.md)
* [Go API](https://pkg.go.dev/github.com/tliron/prudence)
* [FAQ](FAQ.md)


Highlights
----------

* A triple-phase representation process allows for composable, fine-grained, associative control over
  server-side and client-side caching. Reap the full benefits of idempotency in RESTful network
  architectures.
* Prudence's core is written in compiled Go for reliability and performance but allows for interpreted
  JavaScript or TypeScript for your application. This is the right balance between power and productivity.
* Or use [JavaScript Templates (JST)](jst/README.md) to generate HTML by combining design with short
  code scriptlets. And there's sugar.
* Pluggable server-side cache backends. Included is a powerful distributed memory cache that is
  Kuberentes-aware. Or choose backends for [Memcached](https://memcached.org/),
  [Redis](https://redis.io/), etc. Go nuts.
* Schedule jobs using a crontab-like pattern.
* [Extensible](platform/README.md) via the [xprudence tool](xprudence/README.md), which allows you
  to create custom builds of Prudence bundled with the plugins and APIs you need. Even when extended in
  this way Prudence still remains a single, compact executable file.
