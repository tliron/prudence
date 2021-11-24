Prudence: Tutorial
==================

Let's learn by example and build a web application using Prudence.

Table of Contents:

* [The Server](#the-server)
* [The Router](#the-router)
* [A Dynamic Resource](#a-dynamic-resource)
* [Effects](#effects)
* [JavaScript Templates (JST)](#javascript-templates-jst)
* [Rendering](#rendering)
* [Scheduler](#scheduler)
* [Next Steps](#next-steps)

### TypeScript

In this tutorial we will be using JavaScript. However, Prudence also supports
[TypeScript](https://www.typescriptlang.org/). See the
[example](https://github.com/tliron/prudence/tree/main/examples/typescript) to get started.


The Server
----------

Our starting point is the `prudence run` command, which runs JavaScript code in Prudence's
environment. So, let's create a `start.js` file with the simplest code possible:

    prudence.start(new prudence.Server({address: ':8080'}));

And then run it from a terminal:

    prudence run start.js -v

If there were no errors then you should see a log message about the server being up and
running. The program will now block until you kill it, e.g. by pressing CTRL+C.

Let's open another terminal and test it out:

    curl localhost:8080 -v

You should see a successful response with a 200 status code.

### "prudence.start"

This function's argument is either a single startable object or an array of startable objects.
So, you can start several servers at the same time, e.g. to listen on different ports or
interfaces.

Prudence will automatically restart itself if any of the dependent files (JavaScript source
code or others, such as loaded/included files) are changed. To do this it "watches" these files
using filesystem services. To turn this feature off run Prudence with the `--watch=false` flag.
Note that restarting the server(s) does *not* delete any cached representations, even 
if you're using the in-memory cache backend.

### TLS

By default the server is unencrypted HTTP/1.1. To secure the connection ("https:" with
support for HTTP/2) you need to set the certificate and key PEMs, either literally or by
loading them from a file:

    prudence.start(new prudence.Server({
        address: ':8080',
        secure: {
            certificate: prudence.loadString('secret/crt.pem'),
            key: prudence.loadString('secret/key.pem')
        }
    }));

For testing you can set `secure: {}` to automatically create an self-signed certificate.
To see it in the terminal:

    curl --insecure https://localhost:8080 -v

### NCSA

To enable an [NCSA Common log](https://en.wikipedia.org/wiki/Common_Log_Format) run Prudence
with the `--ncsa` flag. You can give it a path to a log file or the special values "stdout"
or "stderr":

    prudence run start.js --ncsa=/var/log/ncsa.log -v

If you have multiple servers you can use a "ncsa" property that will be added as a prefix to
the `--ncsa` filename:

    prudence.start(new prudence.Server({
        address: ':8080',
        ncsa: 'main'
    }));

We have a server running. Now, let's add an application!


The Router
----------

Let's create a directory for our application, `myapp`. You don't have to organize your files
in any particular way, but it's nice to have the application in its own directory. After
all, you might want to have more than one application running on the same server. And then
create a `myapp/router.js` file with this code:

    exports.handler = new prudence.Router({
        name: 'myapp',
        routes: [{
            handler: new prudence.Static({
                root: 'files/'
            })
        }, {
            handler: prudence.defaultNotFound
        }]
    });

The "name" property is optional and only used for logging, which is useful if you have
multiple applications running on the same server. A few Prudence types support this property.

Create a `myapp/files` directory and put any file(s) you want there. Note that the "root" path
is relative to the current JavaScript file's directory. Actually, in Prudence almost all file
references are relative to the current JavaScript file's directory.

Then, edit your `start.js` with this code:

    prudence.start(new prudence.Server({
        address: ':8080',
        handler: require('./myapp/router').handler
    }));

### "exports" and "require"

Prudence is a [CommonJS-style modular](http://www.commonjs.org/specs/modules/1.0/)
JavaScript environment. Simply put, any module can "export" values, including functions,
by placing them in its "exports" global. Other modules can use "require" to access those
exports. Prudence caches the module's exports, so that any module is only ever executed
once.

"require" first looks for the module relative to the current module's path and then relative
to the general path. You can set the general path using the `--path` flag for `prudence`
or via the `PRUDENCE_PATH` environment variable. If an extension is not provided then `.js` will
be added.

The "require" here will give us access to the "exports.handler" from `router.js`.

You should have a running static file server now. You can use `curl` again to test it, or
open a web browser to [`http://localhost:8080`](http://localhost:8080). The Prudence
"Static" handler will, by default, automatically generate an HTML page listing the contents
of the directory.

In the above code we've set three different "handlers". What are these?

### Handlers

Like many web frameworks, Prudence is based on chaining together "handlers" that can modify
or terminate a request. Handlers can also delegate to other handlers based on aspects of the
request (we call this "routing", not to be confused with network routing in the IP protocol!).

The "Server" object only allows for a single handler, so it's very common to set it to be
a "Router" handler. Routers, as you can see above, allow for multiple "routes", each with
its own handler. Each route is attempted *in order*. In this case we are trying to handle
the reqest with a "Static" handler, and if that fails (file not found) it will move on to
the next route, which is Prudence's default 404 Not Found handler.

Note that it's possible to use the same router with multiple servers.

Also note that if none of Prudence's built-in handlers do what you need then you can always
implement a handler directly in JavaScript. A common use case is programmatic redirection,
a.k.a. URL rewriting:

    exports.handler = new prudence.Router({
        name: 'myapp',
        routes: [{
            handler: function(context) {
                const p = context.path.indexOf('/product/');
                if (p != -1) {
                    context.redirect('https://supplier.com/?product=' + context.path.substr(p+9), 301);
                    return true;
                } else {
                    return false;
                }
            }
        }, {
            handler: new prudence.Static({
                root: 'files/'
            })
        }, {
            handler: prudence.defaultNotFound
        }]
    });

Actually, the above solution is not best. It's better to put the handler function in a separate
file and "bind" it, like so:

    exports.handler = new prudence.Router({
        name: 'myapp',
        routes: [{
            handler: bind('./rewrite', 'handler')
        }, {
            handler: new prudence.Static({
                root: 'files/'
            })
        }, {
            handler: prudence.defaultNotFound
        }]
    });

The `rewrite.js` file:

    exports.handler = function(context) {
        const p = context.path.indexOf('/product/');
        if (p != -1) {
            context.redirect('https://supplier.com/?product=' + context.path.substr(p+9), 301);
            return true;
        } else {
            return false;
        }
    };

Why is this better? And what is "bind"?

### "bind"

A "bind" works a lot like a "require": it runs the the JavaScript code and returns the "exports".
The difference is that you cannot use the bound functions in JavaScript. Instead, "bind" prepares
the "exports" for hooking into Prudence's multi-threaded environment so that multiple requests will
be handled simultaneously by the bound code.

If you use "require" instead of "bind" it will still work, but it won't perform as well under load
because Prudence will have to switch to JavaScript's single-threaded execution environment and requests
will have have to wait in line to be handled.

Note that you do not have to use "bind" for Prudence's built-in types, such as "Resource" and
"Router". It is only necessary for JavaScript code.

One consequence of "bind" is that it creates a new execution environment for the bound code. Bound
code will thus not share the same JavaScript globals as other code. To get around this divide you
can use "prudence.globals", which are truly global.

Writing code for a multi-threaded environment is not trivial. You might need to rely on
synchronization techniques for accessing shared data, for example a mutex created by calling
"prudence.mutex". You can even store such a mutex in "prudence.globals". For an example, see our
[`backend.js` file in the hello-world](https://github.com/tliron/prudence/blob/main/examples/hello-world/myapp/person/backend.js).

By the way, another option for improving performance is to write critical handlers in Go. (See the
[extension guide](platform/README.md)). However, as always, avoid premature optimization. A "bound"
JavaScript function will take you very far indeed.

OK, so we've learned how to serve static files. What about dynamic resources?


A Dynamic Resource
------------------

Let's create a `myapp/person` directory and a `myapp/person/resource.js` file with this
code:

    exports.handler = new prudence.Resource({
        facets: {
            paths: '{name}',
            representations: {
                functions: bind('./json')
            }
        }
    });

A "Resource" is similar to a "Router" except it has "facets" instead of "routes". Whereas a
route is an arbitrary handler, a facet handles a request by generating a representation
or otherwise affecting the resource (more on that later). A resource can have multiple facets,
and each facet can have multiple representations. Each representation usually targets one or
more content types. For example, you might have one representation for HTML and another
representation for both JSON and YAML. You can think of the "Resource" as encapsulating state
with all facets making use of the same basic state.

You'll notice that we're using "bind" again, this time without the second argument, which will
bind *all* the exported functions.

### "present"

Now let's create the representation, `myapp/person/json.js`:

    exports.present = function(context) {
        const data = {name: context.variables.name};
        context.writeJson(data);
        context.response.contentType = 'application/json';
    };

The name of this exported function, "present", is required by Prudence. The "functions"
property in `resource.js` expects this and other hook names. (We'll get to the other optional
hooks later on in this tutorial.)

Now, edit your `myapp/router.js` with this code:

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

We've added an additional route before our other two routes, though in this case we've
also given it a "paths" property. Above you've seen that the resource's facet has a "paths"
property, too.

### "paths"

"paths" is used often in Prudence to control routing. The incoming request URL's path is
checked against our "paths" before being handled. "paths" can be a single string, which is
what we used here, or a list of strings, in which case any of the strings in the list can
match (a logical "or"). They can include some special wildcard characters:

* A `*` wildcard in "paths" matches *any* (or no) characters. So our `person/*` in
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
4. The first route's "paths" is `person/*`. This matches! Before calling the handler the
   path is changed to the wildcard's value, which is `linus`.
5. The handler is a resource with a single facet, so let's try it.
6. The first facet's "paths" is `{name}`. This matches! Before selecting a representation
   the wildcard's value is placed in the "name" variable.
7. The facet has only one representation, so we'll choose it. The representation is hooked
   to `json.js`.
8. The `json.js` file doesn't have any of the optional hooks (more on those later), so we'll
   just call its "present" function.
9. "present" sets the content type to JSON and writes JSON to the response using the
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
client-side code! That code is run in clients' browsers, not in Prudence.)


Effects
-------

The "present" hook should not cause server-side changes to the resource. However, there are
three more hooks you can add to your representation to allow for changes, operations, and other
effects.

### "erase"

This hook works with the DELETE verb. The meaning of erasure is up to you: it could be about
deleting data from a database, closing a session, invalidating a cache, etc. Just remember
that DELETE is idempotent: multiple DELETE requests to the same URL should have the same
overall result. A bad example of erasure would thus be decreasing an existing counter, because
several calls to DELETE would result in a different final number.

Yes, we did say *multiple DELETE requests to the same URL*. Though the HTTP specification
doesn't explicitly forbid request bodies in DELETE, many implementations ignore or even discard
the body. It is thus generally not a good idea to take the request body into consideration in
"erase" and stick to the URL and headers.

If your erasure succeeds you must set "context.done" to true. Writing a response body is
optional, and indeed many clients may ignore it, just like they ignore the request body.
Prudence will automatically set the return code to 204 (No Content) if you don't write
anything. If you do want a response body, it might make sense to call "present" to render
the full representation after the erasure.

In addition to setting "context.done" to true you can also set "context.async" to true, which
tells the client that erasure is going to happen soon via status 202 (Accepted). This could be
a useful optimization for improving throughput, because you can respond to the request quickly
and do the actual erasure asynchronously. Of course in some situations it may be important to
do a synchronized erasure. Async example:

    exports.erase = function(context) {
        prudence.go(function() { // runs the function in a thread
            mydb.delete(context.variables.name);
        });
        context.done = true;
        context.async = true;
    }

(Note that setting "context.done" to true will also delete any server-side cache for the
representation. We'll get to that in the caching guide.)

### "modify"

This hook works with the PUT verb. Again, the actual meaning of modification is up to you.
Importantly, modification refers to both creation *and* change. Many people think that PUT is
meant for creation and POST is meant for change (or the other way around). Those people are
wrong: POST has an entirely different use, which we'll get to below.

PUT is special because, like DELETE, it is idempotent. Again, think back to the example of
decreasing a counter: when your "modify" is called multiple times in succession with the same
request the overall result should be the same.

Erasure, of course, is also a kind of modification, and indeed the "modify" hook can also be used
to erase parts of or even the whole resource. For example, an empty request or an empty JSON array
or some other directive can be understood by your "modify" to mean erasure. However, the "erase"
hook is more specialized in that it allows for the "async" response and is sometimes optimized to
ignore the request body. So, it's generally better to use "erase" for erasure if you can. Its
semantics are more optimal for that.

As with "erase" you can set "context.done" to true if the modification happened. There is no
support for "context.async". However, you can optionally set "context.created" to true to let
the client know that the resource was created rather than changed.

Another option is *not* to set "context.done" to true and instead do a 303 (See Other) redirect.
This is often used for web pages, so that if the user refreshes the page then the PUT will not
happen again, possibly unintentionally:

    context.redirect(newUrl, 303);

Though not required, it's often a good idea with PUT to return the modified representation,
essentially what we see in the next GET. So it might make sense to just call "present". Example:

    exports.modify = function(context) {
        if (mydb.update(context.variables.name) == 'new') {
            context.created = true;
        }
        context.done = true;
        exports.present(context);
    }

(Note that setting "context.done" to true will also update the server-side cache for the
representation. We'll get to that in the caching guide.)

### "call"

This hook works with the POST verb. It is the most general-purpose verb, and thus the least
possible to optimize. Indeed, POST is non-idempotent and non-cacheable and should be your last
resort.

"call" is what you use for non-idempotent modifications, such as decreasing a counter. But it
doesn't just have to be about modifications. "call" can run a job, start a workflow, process
a payment, or indeed call an API. It's for any server-side operation on your resource.

OK, so we know how to create a dynamic resource. But there has to be an easier way to
generate HTML other than JavaScript calls to "context.write", right?


JavaScript Templates (JST)
--------------------------

The "present" function in JavaScript gives us a lot of power, but it can be inconvenient
if most of what we're doing is generating HTML. For this reason Prudence comes with an
extensible templating engine, JST, which reverses the order: HTML is the "first-class"
default mode, while JavaScript code has to be explicitly delimited. Additionally, JST comes
with lots of useful sugar.

Let's start simple and add another representation to our `resource.js`:

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

You'll notice that we are now adding a "contentTypes" property to this representation.
This will be matched intelligently against the
[Accept header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept) sent
by the client. "contentTypes" can be a single string or a list of strings (in which case
any of them can match; logical "or"). The representations are matched in order, so the
JSON representation is our default fallback in case HTML does not match.

Now, let's create `myapp/person/html.jst`

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

As you can see, we embed arbitrary JavaScript code using the `<%` and `%>` delimiters.
The first character right after the opening delimiter is used for sugar. In this case
the `<%==` sugar writes the context variable in-place and `<%=` writes any JavaScript
expression, in this case a local number variable.

For more JST sugar see the [JST documentation](jst/README.md). It is even possible to
[extend](platform/README.md#jst-sugar) JST with your own custom sugar.

Behind the scenes the the entire JST file is translated into JavaScript code and wrapped
in an exported "present" function, allowing it to be used with "require" in the same way we
hooked `json.js`.

If you now check the [`http://localhost:8080/person/linus`](http://localhost:8080/person/linus)
URL in your web browser, it will indeed default to this HTML representation, because that's
what web browsers perfer. With curl you need to explicitly ask for HTML:

    curl localhost:8080/person/linus -v --header 'Accept: text/html'

JST also makes it easy to create reusable page templates by capturing content and embedding
representations. For example, here's a `html.jst`:

    <%! 'body' %>
        This is the body
    <%!!%>
    <%& 'template.jst' %>

And `template.jst`:

    <html>
    <head><title>My Site</title></head>
    <body>
        <%== 'body' %>
    </body>
    </html>


Rendering
---------

Prudence allows for pluggable "renderers" that can transform text in various ways, including
rendering markup languages, such as
[Markdown](https://daringfireball.net/projects/markdown/).

Programmatic use of "prudence.render":

    let content = prudence.loadString('readme.md');
    content = prudence.render(content, 'markdown');
    context.write(content);

In JST you can do the same with the "include" sugar, `<%+`:

    <%+ 'readme.md', 'markdown' %>

Or just render an area of the JST with the "render" sugar, `<%^`:

    <%^ 'minhtml' %>
      <div>
        Minimize!
      </div>
    <%^^%>

For a list of all supported renders see the [documentation](render/README.md).


Scheduler
---------

Prudence makes it easy to schedule jobs using a crontab-like pattern. Enable the feature
before calling "prudence.start" in your `start.js`:

    prudence.setScheduler(new prudence.LocalScheduler());

Then you can schedule any function. For example, let's run a function every 10 seconds:

    prudence.schedule('1/10 * * * * *', function() {
        prudence.log.info('scheduled hello!');
    });

Note that you can call "prudence.schedule" at any time, not just in `start.js`.


Next Steps
----------

You're now an expert on all the basics and also very smart and attractive. It is recommended
to continue to the [caching guide](CACHING.md).
