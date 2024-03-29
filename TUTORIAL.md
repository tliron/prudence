Prudence: Tutorial
==================

Let's learn by example and build a web application using Prudence and JavaScript.

As you follow along, you might want to have the
[API documentation](https://prudence.threecrickets.com/assets/typescript/prudence/docs/) open
in a separate tab for reference.

Table of Contents:

* [The Server](#the-server)
* [The Router](#the-router)
* [A Dynamic Resource](#a-dynamic-resource)
* [Effects](#effects)
* [JavaScript Templates (JST)](#javascript-templates-jst)
* [Rendering](#rendering)
* [Scheduler](#scheduler)
* [Next Steps](#next-steps)

Foreward
--------

### Go

Prudence is written in Go and can be used in Go as an imported library.
See [the Go example](https://github.com/tliron/prudence/tree/main/examples/go/).

If you're a 100% Go programmer, you might be wondering if Prudence is for you at all, what with
all this JavaScript and JST "bloat". After all, Go already has pleasantly straightforward and
powerful built-in libraries for building HTTP backends.

Let's be clear: Prudence is truly a Go library first and foremost, with the JavaScript interface
being entirely optional. If you do not use JavaScript then it will never be running in your
program. We chose JavaScript for this tutorial because we wanted to showcase it, but do not
be confused: every feature demonstrated below can be used directly in Go.

What Prudence adds to barebones Go is considerable:

* Comprehensive HTTP content negotiation and handling of conditional requests, all integrated
  with server-side and client-side caching. The semantics of these features are tricky to get
  right and Go does not provide them out of the box. They are key to truly scalable RESTful
  backends and user experiences.
* REST-oriented hiearchical routing based on URL templates, from application to resource to
  facet to representation. Of course if all you need is improved routing then there are libraries
  that do *only* that, such as [gorilla/mux](https://github.com/gorilla/mux). Still, you might
  appreciate Prudence's opinionated take on routing, such as its straightforward support for
  trailing-slash redirection.
* [JavaScript Templates (JST)](#javascript-templates-jst). Yeah yeah, we snuck this in. Of course
  there are many HTML templating libraries for Go, including the built-in
  [html/template](https://pkg.go.dev/html/template). They might satisfy your needs entirely. But we
  personally prefer the power of a full-blown programming language over an anemic
  [DSL](https://en.wikipedia.org/wiki/Domain-specific_language), and we chose JavaScript because
  you're probably already using it as your frontend in-browser language. Note that our JST
  implementation is available as its own library,
  [go-scriptlet](https://github.com/tliron/go-scriptlet), in case you want to use it without the
  rest of Prudence.

So, Go programmer: We suggest reading through the tutorial and keeping in mind that you can do
all of this in Go. You might still appreciate what Prudence brings to the table.

### TypeScript

Do you hate JavaScript's dynamic and loosey-goosey type system? Then use
[TypeScript](https://www.typescriptlang.org/) instead. Prudence comes with TypeScript definitions
for all its JavaScript APIs. See the
[example](https://github.com/tliron/prudence/tree/main/examples/typescript) to get started.


The Server
----------

Our starting point is the `prudence run` command, which runs JavaScript code in Prudence's
environment. So, let's create a `start.js` file with the simplest code possible:

```javascript
prudence.start(new prudence.Server());
```

And then run it from a terminal (with `-v` for more verbose logging):

    prudence run start.js -v

If there were no errors then you should see a log message about the server being up and
running on the default port, 8080. The program will now block until you kill it, e.g. by
pressing CTRL+C.

Let's open another terminal and test it out:

    curl localhost:8080 -v

You should see a successful response with a 200 status code.

### `prudence.start(...)`

This function's argument is either a single startable object or an array of startable objects.
So, you can start several servers at the same time, e.g. to listen on different ports or
interfaces.

Prudence will automatically restart itself if any of the dependent files (JavaScript source
code or others, such as loaded/included files) are changed. To do this it "watches" these files
using filesystem services. To turn this feature off run Prudence with the `--watch=false` flag.
Note that restarting the server(s) this way does *not* delete any cached representations, even 
if you're using the in-memory cache backend.

### Secure the Server

By default the server is unencrypted HTTP/1.1. To secure the connection ("https:" with
support for HTTP/2 clients) you need to set the PEMs for the TLS certificate and key, either
literally or by loading them from a file:

```javascript
prudence.start(new prudence.Server({
    port: 8081,
    tls: {
        certificate: env.loadString('secret/crt.pem'),
        key: env.loadString('secret/key.pem')
    }
}));
```

Note that when using curl with "https:" on a port that is not 443 you will need an extra
argument:

    curl "https://localhost" --cacert secret/crt.pem --connect-to localhost:443:localhost:8081

For testing you can also set `tls: {generate: true}` to generate a self-signed certificate. To
access with curl:

    curl --insecure https://localhost:8080 -v

### NCSA Logging

To enable an [NCSA Common log](https://en.wikipedia.org/wiki/Common_Log_Format) run Prudence
with the `--ncsa` flag. You can give it a path to a log file:

    prudence run start.js --ncsa=/var/log/ncsa.log -v

If you have multiple servers you can use a `ncsaLogFileSuffix` property that will be added as
a suffix to the `--ncsa` filename:

```javascript
prudence.start(new prudence.Server({
    ncsaLogFileSuffix: '-main'
}));
```

Note that if you're logging NCSA to `/dev/` (e.g. `/dev/stderr`) then the suffix is ignored.

We have a server running. Now, let's add an application!


The Router
----------

Let's create a directory for our application, `myapp`. You don't have to organize your files
in any particular way, but it's nice to have the application in its own directory. After
all, you might want to have more than one application running on the same server. See the
[skeleton example](https://github.com/tliron/prudence/tree/main/examples/skeleton) for a
suggestion on how to structure all of your code.

Create a `myapp/myapp.js` file with this code:

```javascript
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{
        handler: new prudence.Static({
            root: 'files/'
            indexes: 'index.html',
            presentDirectories: true
        })
    }, {
        handler: prudence.defaultNotFound
    }]
});
```

The `name` property is optional and only used for logging, which is useful if you have
multiple applications running on the same server. Almost all Prudence types support this property.

Create a `myapp/files` directory and put any file(s) you want there. Note that the "root" path
is relative to the current JavaScript file's directory. Actually, in Prudence almost all file
references are relative to the current JavaScript file's directory.

Then, edit your `start.js` with this code:

```javascript
prudence.start(new prudence.Server({
    handler: require('./myapp/myapp').handler
}));
```

### `exports` and `require`

Prudence is a [CommonJS-style modular](http://www.commonjs.org/specs/modules/1.0/)
JavaScript environment. Simply put, any module can export values, including functions,
by placing them in its `exports` global. Other modules can use `require` to access those
exports. Prudence caches the module's exports, so that any module is only ever executed
once even if it's required multiple times in your program.

`require` first looks for the module relative to the current module's path and then relative
to the general path. You can set the general path using the `--path` flag for `prudence`
or via the `PRUDENCE_PATH` environment variable. If an extension is not provided then `.js` will
be added.

The `require` here will give us access to the `exports.handler` from `myapp.js`, which is a
`Router`.

You should have a running static file server now. You can use `curl` again to test it, or
open a web browser to [`http://localhost:8080`](http://localhost:8080). The Prudence
`Static` handler will, by default, automatically generate an HTML page listing the contents
of the directory (because we set `presentDirectories` to true).

In the above code we've set three different "handlers". What are these?

### Handlers

Like many HTTP frameworks, Prudence is based on chaining together "handlers" that can modify
or terminate a request. Handlers can also delegate to other handlers based on aspects of the
request (we call this "routing", not to be confused with network routing in the IP protocol!).

The `Server` object only allows for a single handler, so it's very common to set it to be
a `Router` handler. Routers, as you can see above, allow for multiple "routes", each with
its own handler. Each route is attempted *in order*. In this case we are trying to handle
the request with a `Static` handler, and if that fails (file not found) it will move on to
the next route, which is Prudence's default 404 Not Found handler.

Note that it's possible to use the same router with multiple servers, so you can serve the
same application on multiple addresses and ports, some secure and some not.

Also note that if none of Prudence's built-in handlers do what you need then you can always
implement a handler directly in JavaScript. A common use case is programmatic redirection,
a.k.a. URL rewriting:

```javascript
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{
        handler: function() {
            const p = this.request.path.indexOf('/product/');
            if (p != -1) {
                this.redirect('https://supplier.com/?product=' + this.request.path.substr(p+9), 301);
                return true;
            } else {
                return false;
            }
        }
    }, {
        handler: new prudence.Static({
            root: 'files/'
            indexes: 'index.html',
            presentDirectories: true
        })
    }, {
        handler: prudence.defaultNotFound
    }]
});
```

Actually, the above solution is not best. It's better to use `bind` instead of `require`
for the server's handler:

```javascript
prudence.start(new prudence.Server({
    handler: bind('./myapp/myapp', 'handler')
}));
```

Why is this better? And what is `bind`?

### `bind`

A `bind` works a lot like a `require`: it runs the the JavaScript code and returns the `exports`.
The difference is that you cannot use the bound functions in JavaScript. Instead, `bind` prepares
the `exports` for hooking into Prudence's multi-threaded Go environment so that multiple requests will
be handled simultaneously by the bound code.

If you use `require` instead of `bind` it will still work, but it won't perform as well under load
because Prudence will have to switch to JavaScript's single-threaded execution environment and requests
will have have to wait in line to be handled.

Note that you do not have to use `bind` everywhere. It's enough to have `bind` at just one place in your
handler chain to support multi-threading, so doing so once in your server is enough.

One consequence of `bind` is that it creates a new JavaScript execution environment for the bound code.
Bound code will thus not share the same JavaScript globals as other code. To get around this separation
you can use `env.variables`, which are truly global across all execution environments in your program.

Writing code for a multi-threaded environment is not trivial. You might need to rely on
synchronization techniques for accessing shared data, for example a mutex created by calling
`util.mutex()` and storing it in `env.globals`. For an example, see our
[`backend.js` file in the hello-world](https://github.com/tliron/prudence/blob/main/examples/hello-world/myapp/person/backend.js).

By the way, another option for improving performance is to write critical handlers in Go. (See the
[extension guide](platform/README.md)). However, as always, avoid premature optimization. A `bound`
JavaScript handler will take you very far indeed.

OK, so we've learned how to serve static files. What about dynamic resources?


A Dynamic Resource
------------------

Let's create a `myapp/person` directory and a `myapp/person/resource.js` file with this
code:

```javascript
exports.handler = new prudence.Resource({
    facets: {
        paths: '{name}',
        representations: {
            functions: bind('./json')
        }
    }
});
```

A `Resource` is similar to a `Router` except it has `facets` instead of `routes`. Whereas a
route is an arbitrary handler, a facet handles a request by generating a representation
of the resource or otherwise [affecting it](#effects). A resource can have multiple facets,
and each facet can have multiple representations. Each representation usually targets one or
more content types. For example, you might have one representation for HTML and another
representation for both JSON and YAML. You can think of the "Resource" as encapsulating state
with all facets making use of the same basic state.

You'll notice that we're using "bind" again, this time without the second argument, which will
bind *all* the exported functions.

In this example we're keeping the directory structure simple, but a general good practice for
large projects is to build a directory structure like so:

Host (servers) -> Apps (routers) -> Resources -> Facets -> Representations

Of course, it's up to you. You can even get away with creating an entire Prudence application in
a single JavaScript file!

### `present`

Now let's create the representation, `myapp/person/json.js`:

```javascript
exports.present = function() {
    const data = {name: this.variables.name};
    this.transcribe(data);
    this.response.contentType = 'application/json';
};
```

The name of this exported function, `present`, is required by Prudence. The `functions`
property in `resource.js` expects this and other hook names. (We'll get to the other optional
hooks later on in this tutorial.)

Now, edit your `myapp/router.js` with this code:

```javascript
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{
        paths: 'person/*',
        handler: require('./person/resource').handler
    }, {
        handler: new prudence.Static({
            root: 'files/'
        })
    }, {
        handler: prudence.defaultNotFound
    }]
});
```

We've added an additional route before our other two routes, though in this case we've
also given it a `paths` property. Above you've seen that the resource's facet has a `paths`
property, too.

### `paths`

`paths` is used often in Prudence to control routing. The incoming request URL's path is
checked against our `paths` before being handled. `paths` can be a single string, which is
what we used here, or a list of strings, in which case any of the strings in the list can
match (a logical "or"). They can include some special wildcard characters:

* A `*` wildcard in `paths` matches *any* (or no) characters. So our `person/*` in
  `router.js` would match all of these paths: `person/linus`, `person/linus/projects`,
  and `person/`. The `*` wildcard does more: upon a successful match it also
  *changes the path to the wildcard's value* before calling the handler. What this means
  for our router is that a URL path such as `person/linus/projects` will turn into
  `linus/projects` before being delegated to the resource. Another way to think of this
  is that we "chopped off" the hardcoded part of the URL.
* A string encased in curly brackets, e.g. `{name}`, is a variable wildcard. It's a bit
  more conservative than `*` in that it will stop matching at a `/` path separator. So,
  `{name}` would match `linus` but *won't* match `linus/projects`. Unlike the `*` wildcard
  the URL is *not* changed. However, there is a different effect: the wildcard's value
  is extracted into a variable. So, our `{name}` in `resource.js` can then be accessed as
  "context.variables.name" in our `json.js`.

### A Complete Request

Our application is more complex now, so let's follow a request one step at a time:

1. A GET request comes to the server with this URL: `http://localhost:8080/person/linus`.
   The path segment of the URL is `person/linus`.
2. The server uses the router as its handler.
3. The router has three routes, so let's try them one at a time.
4. The first route's `paths` is `person/*`. This matches! Before calling the handler the
   path is changed to the wildcard's value, which is `linus`.
5. The handler is a resource with a single facet, so let's try it.
6. The first facet's `paths` is `{name}`. This matches! Before selecting a representation
   the wildcard's value is placed in the "name" variable.
7. The facet has only one representation, so we'll choose it. The representation is hooked
   to `json.js`.
8. The `json.js` file doesn't have any of the optional hooks (more on those later), so we'll
   just call its `present` hook.
9. `present` sets the content type to JSON and writes JSON to the response using the
   "name" variable that was extracted via the path wildcard.

### More JavaScript

You might be wondering at this point what APIs are available for your JavaScript code in
Prudence. Can you use libraries downloaded from [npm](https://www.npmjs.com/)? The answer is
a qualified no. Most of those libraries are designed to work with [Node.js](https://nodejs.org/),
which is a JavaScipt environment that is very different from Prudence's. And some are designed
for web browsers, which are different yet again. Generic JavaScript code will work, but anything
that relies on platform-specific APIs will not.

Prudence provides you with an alternative solution: the ability to use almost any Go library
as-is in JavaScript. There's a growing ecosystem of great Go libraries that can help you write
your application, including database drivers. To learn how to use them see the
[extension guide](platform/README.md#javascript-apis).

(To be clear, you can definitely use [npm](https://www.npmjs.com/) JavaScript libraries in your
*client*-side code! That code is run in clients' browsers, not in Prudence.)


Effects
-------

The `present` hook should not cause server-side changes to the presented resource. However,
there are three more hooks you can add to your representation to allow for changes, operations,
and other effects.

### `erase`

This hook works with the DELETE verb. The meaning of erasure is up to you: it could be about
deleting data from a database, closing a session, invalidating a cache, etc. Just remember
that DELETE is idempotent: multiple DELETE requests to the same URL should have the same
overall result. A bad example of erasure would thus be decreasing an existing counter, because
several calls to DELETE would result in a different final number.

Yes, we did say *multiple DELETE requests to the same URL*. Though the HTTP specification
doesn't explicitly forbid request bodies in DELETE, many implementations ignore or even discard
the body. It is thus generally not a good idea to take the request body into consideration in
`erase` and stick to the URL and headers.

If your erasure succeeds you must set `this.done` to true. Writing a response body is
optional, and indeed many clients may ignore it, just like they ignore the request body.
Prudence will automatically set the return code to 204 (No Content) if you don't write
anything. If you do want a response body, it might make sense to call `present` to render
the full representation after the erasure.

In addition to setting `this.done` to true you can also set `this.async` to true, which
tells the client that erasure is going to happen soon via status 202 (Accepted). This could be
a useful optimization for improving throughput, because you can respond to the request quickly
and do the actual erasure asynchronously. Of course in some situations it may be important to
do a synchronized erasure. Async example:

```javascript
exports.erase = function() {
    util.go(function() { // runs the function in a thread
        mydb.delete(this.variables.name);
    });
    this.done = true;
    this.async = true;
}
```

(Note that setting "this.done" to true will also delete any server-side cache for the
representation. We'll get to that in the caching guide.)

### `modify`

This hook works with the PUT verb. Again, the actual meaning of modification is up to you.
Importantly, modification refers to both creation *and* change. Many people think that PUT is
meant for creation and POST is meant for change (or the other way around). Those people are
wrong: POST has an entirely different use, which we'll get to below.

PUT like DELETE, it is idempotent. Again, think back to the example of decreasing a counter:
when your "modify" is called multiple times in succession with the same request the overall
result should be the same.

Erasure, of course, is also a kind of modification, and indeed the "modify" hook can also be used
to erase parts of or even the whole resource. For example, an empty request or an empty JSON array
or some other directive can be understood by your "modify" to mean erasure. However, the "erase"
hook is more specialized in that it allows for the "async" response and is sometimes optimized to
ignore the request body. So, it's generally better to use "erase" for erasure if you can, because
its semantics are optimized for that.

As with "erase" you can set `context.done` to true if the modification happened. There is no
support for `this.async`. However, you can optionally set "context.created" to true to let
the client know that the resource was created rather than changed.

Another option is *not* to set `this.done` to true and instead do a 303 (See Other) redirect.
This is often used for web pages, so that if the user refreshes the page then the PUT will not
happen again, possibly unintentionally:

```javascript
this.redirect(newUrl, 303);
```

Though not required, it's often a good idea with PUT to return the modified representation,
essentially what we see in the next GET. So it might make sense to just call `present`. Example:

```javascript
exports.modify = function() {
    if (mydb.update(this.variables.name) === 'new') {
        this.created = true;
    }
    this.done = true;
    exports.present.call(this);
};
```

(Note that setting `this.done` to true will also update the server-side cache for the
representation. We'll get to that in the caching guide.)

### `call`

This hook works with the POST verb. It is the most general-purpose verb, and thus the least
possible to optimize. Indeed, POST is non-idempotent and non-cacheable and should be your last
resort.

`call` is what you use for non-idempotent modifications, such as decreasing a counter. But it
doesn't just have to be about modifications. `call` can run a job, start a workflow, process
a payment, or indeed call an API. It's for any server-side operation on your resource.


JavaScript Templates (JST)
--------------------------

OK, so we know how to create a dynamic resource. But there has to be an easier way to
generate HTML other than JavaScript calls to "this.write", right?

The `present` function in JavaScript gives us a lot of power, but it can be inconvenient
if most of what we're doing is generating HTML. For this reason Prudence comes with an
extensible templating engine, JST, which reverses the order: HTML is the "first-class"
default mode, while JavaScript code has to be explicitly delimited. Additionally, JST comes
with lots of useful sugar.

Let's start simple and add another representation to our `resource.js`:

```javascript
exports.handler = new prudence.Resource({
    facets: {
        paths: '{name}',
        representations: [{
            functions: bind('./json')
        }, {
            contentTypes: 'text/html',
            functions: bind('./html.jst')
        }]
    }
});
```

You'll notice that we are now adding a "contentTypes" property to this representation.
This will be matched intelligently against the
[Accept header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept) sent
by the client. "contentTypes" can be a single string or a list of strings (in which case
any of them can match; logical "or"). The representations are matched in order, so the
JSON representation is our default fallback in case HTML does not match.

Now, let's create `myapp/person/html.jst`

```
<!DOCTYPE html>
<html>
<body>
    <h1>
        Name: <%== 'name' %>
    </h1>
    <div>
        Some numbers:
<% for (let i = 0; i < 10; i++) { %>
        <div><%= i %></div>
<% } %>
    </div>
</body>
</html>
```

As you can see, we embed arbitrary JavaScript code using the `<%` and `%>` delimiters.
The first character right after the opening delimiter is used for sugar. In this case
the `<%==` sugar writes the context variable in-place and `<%=` writes any JavaScript
expression, in this case a local number variable.

For more JST sugar see the [JST documentation](jst/README.md). It is even possible to
[extend](platform/README.md#jst-sugar) JST with your own custom sugar.

Behind the scenes the the entire JST file is translated into JavaScript code and wrapped
in an exported `present` function, allowing it to be used with "bind" in the same way we
hooked `json.js`. Note that you can also refer to that `present` function directly:

```javascript
bind('./html.jst', 'present')
```

If you now check the [`http://localhost:8080/person/linus`](http://localhost:8080/person/linus)
URL in your web browser, it will indeed default to this HTML representation, because that's
what web browsers prefer. With curl you need to explicitly ask for HTML:

    curl localhost:8080/person/linus -v --header 'Accept: text/html'

JST also makes it easy to create reusable page templates by capturing content and embedding
representations. For example, here's a `html.jst`:

```
<%! 'body' %>
    This is the body captured into context variable 'body'
<%!!%>
<%& 'template.jst' %>
```

And `template.jst`:

```
<html>
<head><title>My Site</title></head>
<body>
    <%== 'body' %>
</body>
</html>
```


Rendering
---------

Prudence allows for pluggable "renderers" that can transform text in various ways, including
rendering markup languages, such as
[Markdown](https://daringfireball.net/projects/markdown/).

Programmatic use of "prudence.render":

```javascript
let content = prudence.loadString('readme.md');
content = prudence.render(content, 'markdown');
this.write(content);
```

In JST you can do the same with the "include" sugar, `<%+`:

```
<%+ 'readme.md', 'markdown' %>
```

Or just render an area of the JST with the "render" sugar, `<%^`:

```
<%^ 'minhtml' %>
    <div>
    Minimize!
    </div>
<%^^%>
```

For a list of all supported renders see the [documentation](render/README.md).


Scheduler
---------

Prudence makes it easy to schedule jobs using a crontab-like pattern. Enable the feature
before calling "prudence.start" in your `start.js`:

```javascript
prudence.setScheduler(new prudence.LocalScheduler());
```

Then you can schedule any function. For example, let's run a function every 10 seconds:

```javascript
prudence.schedule('1/10 * * * * *', function() {
    prudence.log.info('scheduled hello!');
});
```

Note that you can call "prudence.schedule" at any time, not just in `start.js`.


Next Steps
----------

You're now an expert on all the basics and also very smart and attractive. It is recommended
to continue to the [caching guide](CACHING.md).
