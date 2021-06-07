
// You can send arguments to your program via "prudence run --argument=name=value"

for (var name in prudence.arguments) {
    prudence.log.infof('argument %s = %s', name, prudence.arguments[name]);
}

// Setting a cache is a good idea, even if it's the in-memory cache
// (Though note that the in-memory cache is of course insufficient for large
// and/or distributed applications)

prudence.setCache(new prudence.MemoryCache());

// "prudence.start" can accept a single server or a list of servers
// Multiple servers can share the same handler

prudence.start(new prudence.Server({
    debug: true,
    address: 'localhost:8080',
    handler: require('myapp/router.js').handler,
    // An empty object will generate a self-signed certificate
    secure: (prudence.arguments.secure === 'true') ? {} : null
    /* Or specify:
    secure: {
        certificate: prudence.loadString('secret/server.crt'),
        key: prudence.loadString('secret/server.key')
    }*/
}));
