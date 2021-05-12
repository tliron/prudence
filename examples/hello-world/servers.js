
// We can start several servers at once and they can use the same
// or different handlers

prudence.start([prudence.create({
    type: 'server',
    //name: 'MyPrudence',
    address: '127.0.0.1:8080',
    handler: prudence.import('myapp/router.js')
})]);
