
// "prudence.start" can accept a single server or a list of servers
// Multiple servers can share the same handler

prudence.start(prudence.create({
    type: 'server',
    //name: 'MyPrudence',
    address: 'localhost:8080',
    // protocol: 'http2',
    // secure: true,
    handler: prudence.import('myapp/router.js')
}));
