
prudence.start(
    new prudence.Server({
        address: ':8080',
        handler: bind('./myapp/router', 'handler'),
        debug: true
    })
);
