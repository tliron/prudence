Prudence: FAQ
=============

### Why Go?

[Go](https://golang.org/) is fast replacing both Java and Python in many environments.
It has the advantage of producing very deployable executables that make it easy to containerize
and integrate. Go features garbage collection and easy multi-threading (via lightweight
goroutines), but unlike Python, Ruby, and Perl it is a strictly typed language, which
encourages good programming practices and reduces the chance for bugs.

### JavaScript? Really?

First off, Prudence supports [TypeScript](https://www.typescriptlang.org/), which is a big step
forward for the JavaScript ecosystem. TypeScript is a nice language!

As for the underlying JavaScript, it's probably
[not anyone's favorite](https://archive.org/details/wat_destroyallsoftware), but it's
familiar, mature, standardized (as [ECMAScript](https://en.wikipedia.org/wiki/ECMAScript)), and does
the job. From a certain angle it's essentially the Scheme language (because it has powerful closures
and functions are first class citizens) but with a crusty C syntax.

Our chosen ECMAScript engine is [goja](https://github.com/dop251/goja), which is 100% Go and does
not require any external dependencies. It complies with ECMAScript 5.1, but if you really don't want
to use TypeScript and yet still want a more advanced version of JavaScript you can use tools like
[Babel](https://babeljs.io/) or [swc](https://swc.rs/) to transpile it to 5.1.

### Wasn't Prudence originally in Java?

Prudence was first conceptualized in 2009. Those were different times. It was originally
intended as a framework for using various interpreted languages, including templating languages,
to build RESTful pages and APIs. The emphasis was on doing REST right, allowing tight integration
with server-side and client-side caching. The threading model was highly concurrent, which was
against the trend of Node.js-style single-threadedness. Prudence 1 was written in Java, fueled by
[Restlet](https://github.com/restlet/restlet-framework-java) and
[Hazelcast](https://github.com/hazelcast/hazelcast). It was LGPL-licensed and used Subversion as
its VCS (eventually migrating to git). Do you remember
[Google Code](https://code.google.com/archive/p/savory-prudence/)? It was pretty cool.

The JVM is both complex and complicated, thus much of the work focused not on web technologies
but in wrestling with the JVM platform and its ecosystem. And Prudence was ambitious: it supported
JavaScript, Lua, Clojure, Python, Ruby, and more. The project kept getting bulkier and more
expansive until 2014, when it became version 2, at which point it comprised several projects:
Sincerity, Diligence, Succinct, and Scripturian.

The sprawl was unmanageable and development came to a grinding halt. And so in 2021 the project was
rebooted. Code was rewritten from scratch in Go, Apache-licensed, and with a tighter vision with
essentially the same initial concept. Though considerably more lightweight and easier to manage, the
new Prudence is in some ways more powerful and more flexible than it was before. Writing code from
scratch is a great idea if you can afford the time and effort! In conclusion, I hope we've all learned
lessons from our Java foibles. Again: those we different times.

The original code for Prudence in Java is archived [here](https://github.com/tliron/prudence-java).
