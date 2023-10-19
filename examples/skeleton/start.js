
// Setting a cache (optional) is a good idea, even if it's just the in-memory cache
prudence.setCache(new prudence.MemoryCache());

// Setting a scheduler (optional) allows for running tasks using a crontab-like pattern
prudence.setScheduler(new prudence.LocalScheduler());

prudence.schedule('1/10 * * * * *', function() {
    env.log.info('every 10 seconds');
});

var tls = {
    certificate: os.joinFilePath(__dirname, 'secret', 'server.crt'),
    key: os.joinFilePath(__dirname, 'secret', 'server.key')
};

try {
    // Try to load installed secrets
    const certificate = env.loadString(tls.certificate);
    const key = env.loadString(tls.key);
    env.log.notice('using installed secrets')
    tls.certificate = certificate;
    tls.key = key;
} catch (_) {
    env.log.notice('generating secrets')
    tls.generate = true;
}

const router = bind('./router', 'handler');

// Start two servers with the same router
prudence.start([
    new prudence.Server({
        port: 8080,
        debug: true,
        name: 'Server1', // will be sent in the "Server" header for every response
        ncsaLogFileSuffix: '-server1', // only used when the "--ncsa" flag is set
        handler: router
    }),
    new prudence.Server({
        port: 8081,
        tls: tls,
        debug: true,
        name: 'Server2',
        ncsaLogFileSuffix: '-server2',
        handler: router
    }),
]);
