Prudence: Extension Guide
=========================

Prudence is designed to be extensible in a few ways, detailed here.

Plugins would normally be written in Go, against the APIs in this `platform` package. Check out
the [plugin example](https://github.com/tliron/prudence/tree/main/examples/plugin) to see how it
all works together.

Though you could potentially embed Prudence in your custom Go application, the more common use
would be to customize the `prudence` command to be bundled with your plugins. That's what the
[XPrudence](../xprudence/README.md) tool is for, documented separately.

### A note about versions

From Prudence 1.1.0 and onwards the Prudence `platform` package should maintain its contract
between *minor* versions of Prudence. I.e. extensions written against Prudence 1.1.6 should
work with Prudence 1.1.12. The latter may add more features, but should not remove or change the
functionality of existing ones. In other words, if a breaking change needs to be introduced to
this package then the minor version of Prudence would be bumped. Thus extensions written against
Prudence 1.2.0 would not be guaranteed to work with Prudence 1.1.x.

This discipline will not be maintained for 1.0.x versions. In that early stage we will be
doing more frequent API changes as the platform moves towards maturity and stability.


Plugins
-------

Prudence plugins are [Go modules](https://golang.org/ref/mod), so they must have a `go.mod` file.
However, note that your module name does not have to be URL-based if you don't need to publish it
online (the XPrudence tool lets you simply specify a `--directory` for a plugin). In our included
example we indeed get away with calling it simply "myplugin", which we initialized like so:

    go mod init myplugin

Your module's Go code will likely, at the very least, import this package,
"github.com/tliron/prudence/platform". And it would also have at least one "init()" function to
register your extensions. See below for details.

Otherwise, there are no special requirements. You can, for example, use any package name you want
and import anything else. In our example we will call our package "plugin".


JavaScript APIs
---------------

Prudence's JavaScript engine, [goja](https://github.com/dop251/goja), has excellent integration
with Go code, converting values and functions to and from JavaScript for you. This means that in
many cases you can just hand over normal Go code and not worry about JavaScript specificities.
That said, you can receive and create goja types directly for deeper integration, including support
for constructor functions (JavaScript's "new" keyword). See the discussion at
[Runtime.ToValue](https://pkg.go.dev/github.com/dop251/goja?utm_source=godoc#Runtime.ToValue)
for more information.

Let's create a plugin that exposes the [BadgerDB](https://github.com/dgraph-io/badger) API to
JavaScript:

    package plugin

    import (
        badger "github.com/dgraph-io/badger/v3"
        "github.com/tliron/prudence/platform"
    )

    func init() {
        platform.RegisterAPI("badger", API{})
    }

    func (self API) Open(path string) (*badger.DB, error) {
        return badger.Open(badger.DefaultOptions(path))
    }

That's really all there is to it! "badger.open" will return a database instance and all its
methods and types should work fine in JavaScript, including sophisticated things like passing
JavaScript functions to Go code:

    const db = badger.open(prudence.joinFilePath(__dirname, 'db'));

    exports.counter = function(context) {
        var counter;

        db.update(function(txn) {
            try {
                txn.get('counter').value(function(value) {
                    counter = parseInt(prudence.bytesToString(value));
                    return null;
                });
            } catch (e) {
                counter = 0;
            }

            txn.set('counter', prudence.stringToBytes(counter + 1));
            return null;
        });

        return counter;
    };


Custom Types
------------

Prudence has built-in types like "Server", "Router", "Static", "MemoryCache", etc., and you can add your
own. To do this, you need to register a "create" function:

    import "github.com/tliron/kutil/js"

    func init() {
        platform.RegisterType("MyType", CreateMyType)
    }

    type MyType struct{}

    // platform.CreateFunc signature
    func CreateMyType(config map[string]interface{}, context *js.Context) (interface{}, error) {
        return MyType{}
    }

The "config" argument contains the arbitrary data provided in JavaScript's "new". If not provided
it will be an empty map (not a "nil" value). The "context" argument provides access to the JavaScript
runtime environment in which the object is being created. This is useful especially for calling
"context.Resolve", which will let you process relative URLs in the "config".

Like the JavaScript APIs discussed above, your custom types can really do anything you want them to
do. However, you're likely going to be interacting with the Prudence platform in the following ways.

### Handlers

If your type implements the "rest.Handler" interface then it can be used as a handler anywhere in
Prudence, just like "Router", "Resource", and "Static". Example:

    import "github.com/tliron/prudence/rest"

    type MyType struct{
        message string
    }

    // rest.Handler interface
    func (self MyType) Handle(context *rest.Context) bool {
        context.WriteString(self.message + "\n")
        return true
    }

### Startables

If your type implements the "platform.Startable" interface then it can be used as an argument for
"prudence.start". This is a simple interface that just has "Start" and "Stop" methods. The only
built-in startable in Prudence is "Server".

Note that "prudence.start" expects your "Start" implementation to be *blocking*. It will run it in
a goroutine for you. That means that you likely should not be create another goroutine in "Start".
Example:

    type MyType struct{
        stop chan bool
    }

    // platform.Startable interface
    func (self MyType) Start() error {
        <-self.stop // block until a value is sent
        return nil
    }

    // platform.Startable interface
    func (self MyType) Stop() error {
        stop <- true // send a value (and unblock "Start")
        return nil
    }

### Cache Backends

If your type implements the "platform.CacheBackend" interface then it can be used as an argument
for "prudence.setCacheBackend".

Note that only the "LoadRepresentation" method is expected to be synchronous, meaning that it must
return a "CachedRepresentation" if it exists in the cache. The other methods can (and perhaps should)
be asynchronous, meaning that they can return quickly and do the actual work in the background.
Example using an imaginary database:

    // platform.CacheBackend interface
    func (self MyType) LoadRepresentation(key platform.CacheKey) (*platform.CachedRepresentation, bool) {
        if value, ok := db.Get("rep:" + key); ok {
            return unpackCachedRepresentaiton(value), true
        } else {
            return nil, false
        }
    }

    // platform.CacheBackend interface
    func (self MyType) StoreRepresentation(key platform.CacheKey, cached *platform.CachedRepresentation) {
        go func() {
            db.Set("rep:" + key, packCachedRepresentation(cached))
            for _, name := range cached.Groups {
                db.AddToList("grp:" + name, key)
            }
        }()
    }

    // platform.CacheBackend interface
    func (self MyType) DeleteRepresentation(key platform.CacheKey) {
        go func() {
            db.Delete("rep:" + key)
        }()
    }

    // platform.CacheBackend interface
    func (self MyType) DeleteGroup(name platform.CacheKey) {
        go func() {
            if list, ok := db.GetList("grp:" + name); ok {
                for _, key := range list {
                    db.Delete("rep:" + key)
                }
                db.Delete("grp: " + name)
            }
        }()
    }


JST Sugar
---------

If the built-in [JavaScript Template sugar](../jst/README.md) is not sweet enough for you then
you can add your own.

Your custom tag is registered on a prefix, which is a string that will be checked against what
immediately follows the `<%` opening delimiter. Note that not only must it be unique so that it
won't overlap with other tags, but also that it should be unambiguous. Thus you shouldn't register
both the the `-` and the `->` prefixes because the former is included in the latter.

Your tag implementation has two arguments, a "JSTContext" and the raw text between the two JST
delimiters (which includes your prefix). Your implementation can do anything, but what it most
likely will do is write JavaScript source code into the context. Remember that this source
code is eventually integrated in-place into the JST, which is in the end one big "present" hook
function.

The returned value is usually "false", which means that Prudence will swallow the trailing newline
character just after the tag's end delimiter. This is what we want with most tags, as it avoids
filling your output with empty lines. However, you can return "true" to disable this, which is what
the "expression" sugar, `<%=`, does. Also note that the user can explicitly disable this effect by
putting a `/` just before the end delimiter: `/%>`.

Example:

    func init() {
        platform.RegisterTag("~", EncodeInBed)
    }

    // platform.HandleTagFunc signature
    func EncodeInBed(context *platform.JSTContext, code string) bool {
        code = code[1:]
        context.WriteLiteral(strings.TrimSpace(code) + " in bed")
        return false
    }

And then using it in JST:

    <div>
        <%~ I like to watch TV %>
    </div>


Renderers
---------

The Prudence renderer API is quite straightforward: it accepts text as input and returns text
as output. What the [renderer](../render/README.md) actually does, of course, can be quite
sophisticated. It could be an entire language implementation. Here's a trivial example:

    import "github.com/tliron/kutil/js"

    func init() {
        platform.RegisterRenderer("doublespace", RenderDoubleSpace)
    }

    // platform.RenderFunc signature
    func RenderDoubleSpace(content string, context *js.Context) (string, error) {
        return strings.ReplaceAll(context, " ", "  "), nil
    }

Note that the JavaScript context is provided as an argument. This is to allow sophisticated
renderers to integrate with the resolver, module, etc.
