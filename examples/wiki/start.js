
prudence.setCache(new prudence.MemoryCache());

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

prudence.start([
    new prudence.Server({
        port: 8080,
        debug: true,
        name: 'Server1',
        ncsaLogFileSuffix: '-server1',
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
