
// Setting a cache is a good idea, even if it's the in-memory cache
// (Though note that the in-memory cache is of course insufficient for large
// and/or distributed applications)

prudence.setCache(new prudence.MemoryCache());

// "prudence.start" can accept a single server or a list of servers
// Multiple servers can share the same handler

prudence.start(new prudence.Server({
    //name: 'MyPrudence',
    address: 'localhost:8080',
    // protocol: 'http2',
    // secure: true,
    debug: true,
    handler: require('myapp/router.js').handler
}));
