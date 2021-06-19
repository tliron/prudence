Prudence: Caching Guide
=======================

Make sure you're up-to-speed with [the basics](TUTORIAL.md) first!

Table of Contents:

* [Server-Side Caching](#server-side-caching)
* [Client-Side Caching](#client-side-caching)


Server-Side Caching
-------------------

Server-side caching is extremely important for real-world web sites. Even a one-second
cache on your pages can ensure that you'll be able to handle sudden spikes in usage without
hammering your backend.

Let's set up a cache. Add this to your `start.js` *before* you create and start the
server:

    prudence.setCache(new prudence.MemoryCache());

This in-memory cache will suffice for testing and can also be great for smaller web
sites. However, large applications will likely need a distributed cache backend, such
as [Memcached](https://memcached.org/) or [Redis](https://redis.io/). It is also possible
to set up Prudence with tiered caching, so that the faster in-memory cache will be
preferred to the distributed one.

### Cache Duration

Let's enable caching for our `html.jst` representation. You can just add this little
sugar anywhere in the file:

    <%* 10 %>

The `<%*` sugar sets the cache duration to the numeric expression (it's in seconds).
This means that the first time any user requests the URL it will be stored in the
cache. For the next 10 seconds any subsequent requests will use that cached value
instead of regenerating the representation. After those 10 seconds pass any new request
will cause a new representation to be generated, which again will be cached for 10
seconds.

### Coordinating with Clients

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

### "construct"

Now, we remember that our `html.jst` is just one big "present" function, and it's
fine to configure caching there. However, there is a better place to put it. Let's
edit our `json.js` file and add an additional hook function, "construct":

    exports.construct = function(context) {
        context.cacheKey = 'myapp.person.' + context.variables.name;
        context.response.contentType = 'application/json';
    };

You'll notice that we were already setting the content type in the "present" function.
However, now that we have a "construct" function it is the better place for it, so
you can delete that line from "present":

    exports.present = function(context) {
        const data = {name: context.variables.name};
        context.writeJson(data);
        context.cacheDuration = 5;
    };

As for "cacheDuration", you can set it anywhere. However, it might make most sense to
set it in "present", because that's were we usually retrieve our data, the contents of
which may affect our decision about how long it should be cached, if at all.

The "construct" hook is very powerful. If it exists, it is called by Prudence *before*
trying the cache. So, this is where we can set the parameters that tell Prudence how
to retrieve from (and store in) the cache. That's why we need to set the content type
here: Prudence stores each content type as a separate cache entry (obviously, because
they would be different representations).

Also note that "construct" is called before the "erase", "modify", and "call" hooks.

### Cache Keys

We've' also modified the cache key in our "construct". By default the cache key is the
complete URL, which is a sensible default, but might not be the most efficient.

Consider a situation in which this page has multiple, equivalent URLs. For example,
perhaps the site is registered under multiple domain names. If we stick to the default
cache key then we won't be reusing the cache between those domain names. Thus Prudence
lets you customize the cache key as you see fit. You can strip the domain from the URL
or otherwise create any custom key that makes sense to you, which is what we did in
this example.

You just have to be careful that your cache key scheme won't conflict with others,
otherwise you might be mixing cache entries from other parts of the application, or
indeed from other applications using the same cache backend.

### JST

You might be wondering how we can add a "construct" hook when using a JST file, which
only has a "present" function. Instead of the "functions" property we can use properties
named for the individual hooks. `resource.js` could look something like this:

    exports.handler = new prudence.Resource({
        facets: {
            paths: '{name}',
            representations: {
                contentTypes: 'text/html',
                construct: bind('./html', 'construct'),
                present: bind('./html.jst', 'present')
            }
        }
    });

Note that you don't have to use "require" and even an inline function would work:

    construct: function(context) { context.cacheKey = 'person'; }

### Cache Composition

The JST "embed" sugar might appear to work like the "include" suger, however it is far
more powerful:

    <%& 'list.jst' %>

It is used to insert not a raw file but another representation. This means calling the
"present" hook, and indeed also calling "construct" and "describe" if it has them. This
is useful not just for making your code more modular, but also for creating a more
fine-grained caching scheme, because that other representation may also be cached, indeed
with its own cache key and cache duration. Thus, if many different pages use that same
building block they might not have to regenerate it each time.

Because JST files only have "present" and do not have a "construct", the sugar allows you
to optionally add a cache key. Note that this is the key for the embedded representation,
not the containing one:

    <%& 'list.jst', 'person.list.' + context.variables.name %>

### Cache Groups

Deciding on a good cache key scheme can go a long way towards helping your application scale.
However, aggressive caching can introduce data inconsistency. For example, imagine a resource
with several facets, each having several representations, and all are cached. Now a client
sends a DELETE request to one representation, where you have an "erase" hook. Prudence will make
sure to delete the cache entry for that particular representation. But what about all the other
representations? Normally, they stay in the cache until they expire, thus potentially presenting
out-of-date data to clients.

This might not be a problem for your application. But if it is, Prudence provides a powerful
feature to tackle it: cache groups. These are strings that can be assigned to a cache entry in
*addition* to its cache key. They are not used for retrieving cache entries, only for deleting
them to ensure that there is no out-of-date data.

You can add any number of cache groups in "construct":

    exports.construct = function(context) {
        ...
        context.cacheGroups.push('person.resource' + context.variables.name);
    };

And then invalidate the group anywhere in your code (usually in "erase", "modify", or "call"
hooks):

    exports.modify = function(context) {
        ...
        context.done = true;
        prudence.invalidateCacheGroup('person.resource' + context.variables.name);
    }

If you assign the same cache group to all your resource's representations then all of them could
be deleted from the cache with a single call!

Note that as powerful as it is, the feature does come at a cost. For it to be efficient the
cache backend need some sort of indexed querying technology, allowing it to quickly find all cache
entries for a group. Such indexes could take up a lot of precious space that might be better used
for more cache entries. Thus, as always, be careful not to prematurely optimize and to profile your
application's specific behavior under high load.


Client-Side Caching
-------------------

Server-side caching is crucial to making sure your server can scale. But clients can
help, too, by telling the server what they have in *their* cache. They do this by
sending "conditional requests", which can result in the server telling them to continue
using the cached value they have. This means that the server doesn't have to generate
a new representation or even send the cached-on-the-server representation. So, this
saves bandwidth. Also, it improves the user experience as they don't have to download
and re-process the representation.

### "describe"

There is another optional hook just for optimizing his functionality: "describe". Let's
add a "describe" function to our `json.js`:

    exports.describe = function(context) {
        context.signature = prudence.hash(context.cacheKey);
    };

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

### HEAD

For HTTP HEAD requests "describe" is called but "present" *isn't* called. This might
seem like a small (and obvious) optimization, but it can go a long way towards improving
scalability in environments that rely on HEAD.

Also note that HEAD, like GET, still goes through server-side caching. With HEAD, though,
only the headers are written to the response and the cached body is ignored.


A Complete Request
------------------

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
