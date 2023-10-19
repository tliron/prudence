Prudence
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/prudence.svg)](https://github.com/tliron/prudence/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/prudence.svg)](https://pkg.go.dev/github.com/tliron/prudence)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/prudence)](https://goreportcard.com/report/github.com/tliron/prudence)

An opinionated lightweight web framework built for scale, featuring integrated
[RESTful](https://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm) server-side and
client-side caching. Suitable for frontend user interfaces, backend APIs, and full-stack combos.

Write your code in Go, JavaScript, TypeScript, or a combination of programming languages.
Choose the best programming language per need. Balance performance and productivity.
JavaScript/TypeScript are especially useful for bootstrapping your application.

Prudence is distributed as a Go library, as well as a single, compact, customizable executable file with
no external dependencies.

And it's fun! Through rigorous benchmarks conducted in our good-times laboratory we found Prudence to be
a zillion times more fun than competing products from leading brand names.

[![Download](assets/media/download.png "Download")](https://github.com/tliron/prudence/releases)


Highlights
----------

* Structure your code as a routable hierarchy of apps, resources, facets, and representations.
* Rich support for HTTP content negotiation according to format, language, and compression algorithms.
* Use JavaScript Templates (JST) to generate HTML and other textual formats by combining design with short
  code scriptlets. And there's sugar.
* Maximize efficiency via a triple-phase representation process allowing for composable, fine-grained,
  associative control over server-side and client-side caching. *Killer feature!*
* Pluggable server-side cache backends, such as the included Kubernetes-aware distributed memory cache.
  Or choose backends for [Memcached](https://memcached.org/), [Redis](https://redis.io/), etc.
  Or create a tiered cache combining several backends.
* Schedule jobs using a crontab-like pattern.
* [Extensible](platform/README.md) via the [xprudence tool](xprudence/README.md), which allows you
  to create custom builds of Prudence bundled with the plugins and APIs you need. Even when extended in
  this way Prudence still remains a single, compact executable file.


Documentation
-------------

* [Tutorial](TUTORIAL.md)
* [Caching guide](CACHING.md)
* [Examples](examples/README.md)
* [JavaScript/TypeScript API](https://prudence.threecrickets.com/assets/typescript/prudence/docs/)
* [JavaScript Templates (JST)](https://github.com/tliron/go-scriptlet/blob/main/jst/README.md)
* [Renderers](render/README.md)
* [Extension guide](platform/README.md)
* [xprudence](xprudence/README.md)
* [Go API](https://pkg.go.dev/github.com/tliron/prudence)
* [FAQ](FAQ.md)
