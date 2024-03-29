
// You can send arguments to your program via "prudence run --argument=name=value"

for (let name in prudence.arguments) {
    jst.log.infof('argument: %s = %s', name, prudence.arguments[name]);
}

// Setting a cache is a good idea, even if it's just the in-memory cache
// (Though note that the in-memory cache is of course insufficient for large
// and/or distributed applications)

if (prudence.arguments.cache === 'distributed')
    prudence.setCache(new prudence.DistributedCache({
        local: new prudence.MemoryCache(),
        kubernetes: {
            namespace: 'workspace',
            selector: 'app.kubernetes.io/instance=prudence-hello-world'
        }
    }));
else if (prudence.arguments.cache === 'tiered')
    prudence.setCache(new prudence.TieredCache({
        caches: [
            new prudence.MemoryCache(), // first tier
            new prudence.MapCache()     // second tier
        ]
    }));
else
    prudence.setCache(new prudence.MemoryCache());

// Setting a scheduler is optional
// It allows for running tasks using a crontab-like pattern

prudence.setScheduler(new prudence.LocalScheduler());

prudence.schedule('1/10 * * * * *', function() {
    jst.log.info('scheduled hello!');
});

// "prudence.start" can accept a single server or a list of servers
// Multiple servers can share the same handler
// Note if the the cache and scheduler are startables then they will be implicitly started here

prudence.start([
    new prudence.Server({
        address: ':8080',
        handler: bind('./myapp/router', 'handler'),
        secure: (prudence.arguments.secure === 'true') ? {} : null, // an empty object will generate a self-signed certificate
        /* Full "secure" example:
        secure: {
            certificate: prudence.loadString('secret/server.crt'),
            key: prudence.loadString('secret/server.key')
        },*/
        debug: true,
        ncsaLogFilePrefix: '8080-' // only used when the "--ncsa" flag is set
    }),
    new prudence.Server({
        address: ':8081',
        ncsaLogFilePrefix: '8081-',
        handler: bind('./myapp/router', 'handler')
    })
]);
