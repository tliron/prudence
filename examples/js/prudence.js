
// TODO: multiple servers/listeners?

var server = prudence.create({
    type: 'server',
    //name: 'MyPrudence',
    address: '127.0.0.1:8080',
    handler: prudence.import('app/router.js')
});

server.start();
