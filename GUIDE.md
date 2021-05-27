Prudence: Usage Guide
=====================

Let's learn by example and build a web application using Prudence.


The Server
----------

Our starting point is the `prudence run` command, which runs JavaScript code in Prudence's
environment. So, let's create a `start.js` file with the simplest code possible:

    prudence.start(prudence.create({
        type: 'server',
        address: 'localhost:8080'
    }));

And then run it from a terminal:

    prudence run start.js -v

If there were no errors then you should see a log message about the server being up and
running. The command will now block until you kill it, e.g. by pressing CTRL+C.

Let's open another terminal and test it out:

    curl localhost:8080 -v

You should see a successful response with a 200 status code.

Some more detail:

* The "prudence.create" function's argument is an object that at the very least has a
  "type" property. Prudence comes with various essential types, such as the "server" used here,
  and can be extended with your own custom types. The return value is the object you created.
* The "prudence.start" function's argument is either a single startable object or a list of
  startable objects. So, you can start several servers at the same time, e.g. to listen on
  different ports or interfaces.

We have a server running. Now, let's add an application!


The Router
----------

Let's create a directory for our application, `myapp`. You don't have to organize your files
in any particular way, but it's nice to have the application in its own directory. After
all, you might want to have more than one applications running on the same server. And then
create a `myapp/router.js` file with this code:

    prudence.create({
        type: 'router',
        name: 'myapp',
        routes: [{
            handler: prudence.create({
                type: 'static',
                root: 'files/'
            })
        }, {
            handler: prudence.defaultNotFound
        }]
    });

Create a directory `myapp/files` and put any file(s) you want there. (The "root" path is
relative to the current file's directory. Actually, in Prudence almost all file references
are relative to the current file's directory.)

Then, edit your `start.js` with this code:

    prudence.start(prudence.create({
        type: 'server',
        address: 'localhost:8080',
        handler: prudence.run('myapp/router.js')
    }));

If you run `prudence run start.js` again, you should have a running static file server.
You can use `curl` again to test it, or open a web browser to
[`http://localhost:8080`](http://localhost:8080). The Prudence "static" handler will, by
default, automatically generate an HTML page listing the contents of the directory.

In the above code we've set three different "handlers". What are these?

Like many web frameworks, Prudence is based on chaining together "handlers" that attempt to
do something with each incoming request. Handlers can delegate to other handlers based on
certain aspects of the request (we call this "routing", not to be confused with network
routing in the IP protocol!) and may affect certain aspects of the request or the response
along the way. Or, they can declare the handling complete and terminate there (e.g. if they
generated a complete response).

The "server" object only allows for a single handler, so it's very common to set it to be
a "router" handler. Routers, as you can see above, allow for multiple "routes", each with
its own handler. Each route is attempted *in order*. In this case we are trying to handle
the reqest with a "static" handler, and if that fails (file not found) it will move on to
the next route, which is Prudence's default 404 Not Found handler.

Some more detail:

* The "prudence.run" function runs a JavaScript file and returns the *last referenced value*.
  In this case, the only code in `router.js` is a call to "prudence.create" to create a
  router, so that returned router becomes the last value and is in turn returned from
  "prudence.run". Note that this is only one way Prudence lets you use other JavaScript
  code and we'll learn more below.
* The "name" property is optional and only used for logging, so that if you have multiple
  applications running on the same server then you'd more easily be able to read the logs.

We've learned how to server static files. What about dynamic resources?


A Dynamic Resource
------------------

Let's create a `myapp/person` directory and a `myapp/person/resource.js` file with this
code:

    prudence.create({
        type: 'resource',
        facets: {
            paths: '{name}',
            representations: {
                functions: prudence.require('json.js')
            }
        }
    });

A "resource" is similar to a router except it has "facets" instead of "routes". Whereas a
route is an arbitrary handler, a facet handles a request by generating a representation
and/or updating the resource. A resource can have multiple facets, and each facet can have
multiple representations. Each representation usually targets one or more content types.
For example, you might have one representation for HTML and another representation for both
JSON and YAML. This feature is provided for as a convenience to allow you to better separate
and organize code.

Now let's create the representation, `myapp/person/json.js`:

    function present(context) {
        context.log.info('present');
        const data = {name: context.variables.name};
        context.write(JSON.stringify(data));
        context.contentType = 'application/json';
    }

The name of this function, "present", is required by Prudence. The "functions" property in
the representation in `resource.js` uses a call to "prudence.require" to import it (and any
other hooks) from another file. (We'll get to the other optional hooks later on in this guide.)

You might be wondering at this point what APIs are available for your JavaScript code. Can
you use libraries downloaded from [npm](https://www.npmjs.com/)? The answer is a qualified no.
Most of those libraries are designed to work with [Node.js](https://nodejs.org/), which is a
JavaScipt environment that is very different from Prudence's. And some are designed for web
browsers, which are different yet again. Generic JavaScript code will work, but anything that
relies on platform-specific APIs will not.

Prudence provides you with an alternative solution: the ability to use almost any Go library
as-is in JavaScript. There's a growing ecosystem of great Go libraries that can help you write
your application, including database drivers. To learn how to use them see the
[extension guide](platform/README.md).

Now, edit your `myapp/router.js` with this code:

    prudence.create({
        type: 'router',
        name: 'myapp',
        routes: [{
            paths: 'person/*',
            handler: prudence.run('person/resource.js')
        }, {
            handler: prudence.create({
                type: 'static',
                root: 'files/'
            })
        }, {
            handler: prudence.defaultNotFound
        }]
    });

We've added an additional route before our other two routes, though in this case we've
also given it a "paths" property. Above you've seen that the resource's facet has a "paths"
property, too.

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
7. The facet has only representation, so we'll choose it. The representation is hooked to
   `json.js`.
8. The `json.js` file doesn't have any of the optional hooks, so we'll just call its
   "present" function.
9. "present" sets the content type to JSON and writes JSON to the response using the
   "name" variable that was extracted via the path wildcard.

Some more detail:

* The "prudence.require" function runs a JavaScript file and returns the global object.
  The "functions" property in the representation will then look for specifically-named
  hooks. Note that unlike "prudence.require" caches its results, so that the code will
  only ever be run once in your program's lifetime. This is different from "prudence.run"
  which will always run the file when called.

OK, so we know how to create a dynamic resource. But there has to be an easier way to
generate HTML other than JavaScript calls to "context.write", right?


JavaScript Templates (JST)
--------------------------

The "present" function in JavaScript gives us a lot of power, but it can be inconvenient
if most of what we're doing is generating HTML. For this reason Prudence comes with an
extensible templating engine, JST, which reverses the order: HTML is the "first-class"
default mode, while JavaScript code has to be explicitly delimited. Additionally, JST comes
with lots of useful sugar for common tasks.

Let's start simple and add another representation to our `resource.js`:

    prudence.create({
        type: 'resource',
        facets: {
            paths: '{name}',
            representations: [{
                contentTypes: 'text/html',
                functions: prudence.require('html.jst')
            }, {
                functions: prudence.require('json.js')
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
            Name: <%= context.variables.name %>
        </h1>
        <div>
            Some numbers:
    <% for (var i = 0; i < 10; i++) { %>
            <div><%= i %></div>
    <% } %>
        </div>
    </body>
    </html>

As you can see, we embed arbitrary JavaScript code using the `<%` and `%>` delimiters.
The first character right after the opening delimiter is used for sugar. In this case
the `<%=` sugar will simply write the JavaScript expression in-place. For more JST sugar
see the [JST documentation](jst/README.md). It is even possible to extend JST with your
own custom sugar.

Behind the scenes the the entire JST file is translated into JavaScript code and
wrapped in a "present" function, allowing it to be used with "prudence.require" in the same
way we hooked `json.js`.

If you check the [`http://localhost:8080/person/linus`](http://localhost:8080/person/linus)
URL in your web browser, it will indeed default to this HTML representation, because that's
what web browsers perfer. In the CLI you need to explicitly ask for HTML:

    curl localhost:8080/person/linus -v --header 'Accept: text/html'


Rendering
---------

Prudence allows for pluggable "renderers" that can transform text in various ways, including
rendering markup languages, such as
[Markdown](https://daringfireball.net/projects/markdown/).

Programmatic use of "prudence.render":

    var content = prudence.load('readme.md');
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

This covers all the basics. Let's move on to more advanced Prudence functionality.


Server-Side Caching
-------------------

Server-side caching is extremely important for real-world web sites. Even a one-second
cache on your pages can ensure that you'll be handle sudden spikes in usage without
hammering your backend.

Let's set up a cache. Add this to your `start.js` *before* you create and start the
server:

    prudence.setCache(prudence.create({
        type: 'cache.memory'
    }));

This in-memory cache will suffice for testing and can also be great for smaller web
sites. However, large applications will likely need a distributed cache backend, such
as [Memcached](https://memcached.org/) or [Redis](https://redis.io/). It is also possible
to set up Prudence with tiered caching, so that the faster in-memory cache will be
preferred to the distributed one.

Let's enable caching for our `html.jst` representation. You can just add this little
sugar anywhere in the file:

    <%* 10 %>

The `<%*` sugar sets the cache duration to the numeric expression (it's in seconds).
This means that the first time any user requests the URL it will be stored in the
cache. For the next 10 seconds any subsequent requests will use that cached value
instead of regenerating the representation. After those 10 seconds pass any new request
will cause a new representation to be generated, which again will be cached for 10
seconds.

But Prudence does something more sophisticated here. Within that 10 second span we
already know that the representation is not going to change. So, Prudence tells
each individual client to use its *local* client-side cache for that time span.

You can see this in action in your browser. If you're using Firefox or Chrome, go to
[`localhost:8080/person/linus`](localhost:8080/person/linus), press F12 to open
developer tools, and select the "network" tab. Refresh the page several times quickly.
You'll see that the first request receives a 200 status code from the server, meaning
that it received a full representation. But, for the next 10 seconds every request
will show a 304 status code, which means "Not Modified".

Note that setting "cacheDuration" to a negative number has a special meeting: it means
that not only are we not caching on the server (like a zero "cacheDuration"), but also
that we don't want to the client to cache, too. This is sometimes important for
security reasons, i.e. the content contains sensitive information that we'd rather not
be stored anywhere.

We'll discuss client-side caching in more detail in the next section.

Now, we remember that our `html.jst` is just one big "present" function, and it's
fine to configure caching there. However, there is a better place to put it. Let's
edit our `json.js` file and add an additional hook function, "construct":

    function construct(context) {
        context.log.info('construct');
        context.cacheKey = 'myapp.person.' + context.variables.name;
        context.contentType = 'application/json';
    }

You'll notice that we were already setting the content type in the "present" function.
However, now that we have a "construct" function it is the better place for it, so
you can delete that line from "present":

    function present(context) {
        context.log.info('present');
        const data = {name: context.variables.name};
        context.write(JSON.stringify(data));
        context.cacheDuration = 5;
    }

As for "cacheDuration", you can set it anywhere. However, it might make most sense to
set it in "present", because that's were we often retrieve our data, which may inform
out decision about how long it should be cached, if at all.

The "construct" hook is very powerful. If it exists, it is called by Prudence *before*
trying the cache. So, this is where we can set the parameters that tell Prudence how
to retrieve from (and store in) the cache. That's why we need to set the content type
here: Prudence stores each content type as a separate cache entry (obviously, because
they would be different representations).

We can also modify the cache key here. By default the cache key is the complete URI,
which is a sensible default, but might not be the most efficient.

Consider a situation in which this page has multiple, equivalent URLs. For example,
perhaps the site is registered under multiple domain names. If we stick to the default
cache key then we won't be reusing the cache between those domain names. Thus Prudence
lets you customize the cache key as you see fit. You can strip the domain from the URL
or otherwise create any custom key that makes sense to you, which is what we did in
this example.

You just have to be careful that your cache key scheme won't conflict with others,
otherwise you might be mixing cache entries from other parts of the application, or
indeed from other applications using the same cache backend.

You might be wondering how we can add a "construct" hook when using a JST file, which
only has a "present" function. Instead of the "functions" property we can use properties
named for the individual hooks. `resource.js` could look something like this:

    prudence.create({
        type: 'resource',
        facets: {
            paths: '{name}',
            representations: {
                contentTypes: 'text/html',
                construct: prudence.require('html.js').construct,
                present: prudence.require('html.jst').present,
                runtime: runtime
            }
        }
    });

Note that when specifying individual hooks we also have to set the "runtime" property
to the current runtime: `runtime: runtime`. This allows Prudence to properly call
JavaScript code. (You don't need to set "runtime" when using the "functions" property.)

Also note that you don't have to use "prudence.require" and even an inline function would
work:

    present: function(context) { context.write('Hello'); }


Client-Side Caching
-------------------

Server-side caching is crucial to making sure your server can scale. But clients can
help, too, by telling the server what they have in *their* cache. They do this by
sending "conditional requests", which can result in the server telling them to continue
using the cached value they have. This means that the server doesn't have to generate
a new representation or even send the cached-on-the-server representation. So, this
saves bandwidth. Also, it improves the user experience as they don't have to download
and re-process the representation.

There is another optional hook just for optimizing his functionality: "describe". Let's
add a "describe" function to our `json.js`:

    function describe(context) {
        context.log.info('describe');
        context.signature = prudence.hash(context.cacheKey);
    }

The main responsibiliy of the "describe" hook is to set either "signature" and/or
"timestamp". The signature is sent to the client as an
[ETag](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag) during normal
requests, and compared against the signature the client already has during conditional
requests. The reason Prudence has separate hooks for "describe" from "present" is exactly
because we might not need to call "present" at all in case the signatures match. Thus,
for this to be a meaningful optimization "describe" must be *much less expensive* than
"present". This can be challenging and even impossible to achieve in some cases. It all
depends on whether there is a cheap way to get a signature from your data backend.

Note that clients normally cache a resource with the *complete URL* as the key, which
includes all the `?` query parameters. This means that any change to a query parameter
will *not* use a chached representation for different query parameters. Thus adding a
query parameter, which is not processed by the server, is sometimes used as a way to
"punch" through the cache. It works, but be aware that it will leave the other, unused
representations still cached on the client, which could be a waste of space.

Also note that there is no time limit on client-side caching. If the signature never
changes, then conditional requests (from those clients who have the representation
cached) will always return a 304: Not Modified.

In this trivial example we are reusing our (server-side) cache key that we created in
in the "construct" function. We can do this only because we know for sure that the
resulting representation depends entirely and only on that "name" variable.

Now that we covered all three hooks, let's follow a request through them:

1. Let's assume that the client already has a cached representation together with
   the signature we provided from a previous request.
2. So now the client sends a conditional request, with that signature, that is routed
   to our resource's only facet and then to our fallback JSON representation.
3. The "construct" hook is called first. Prudence uses the cache key (and content type)
   to check the server-side cache. If it's cached then we can check the cached
   signature against the signature provided by the client. If the signatures match
   we stop here, 304: Not Modified. If they don't match, we send the client our
   server-side cache entry because the client does not have it.
4. What if there was no hit on the server-side cache? So now Prudence calls the
   "describe" hook which provides us with our signature. If the signature matches the
   client's we stop here, 304: Not Modified. If they don't match, Prudence continues to
   the "present" hook.
5. The "present" hook generates a completely new representation. We return it to the
   client together with the signature we got from the call to "describe".
6. Is our cacheDuration > 0? If so, we store this new representation in the cache using
   the key set in "construct".

If you've followed the above carefully you can see that in "present" you can always
assume that "describe" was previously called and that in "describe" you can always assume
that "construct" was previously called.
