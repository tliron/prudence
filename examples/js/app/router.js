
prudence.create({
    type: 'router',
    name: 'myapp',
    routes: [{
        paths: [ 'persons/*' ],
        handler: prudence.import('person/resource.js')
    }, {
        paths: [ 'static/*' ],
        handler: prudence.create({
            type: 'static',
            root: 'static/',
            indexes: [ 'index.html' ]
        })
    }, {
        // Default handler
        handler: prudence.defaultNotFound
    }]
});
