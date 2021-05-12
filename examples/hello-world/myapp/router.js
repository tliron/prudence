
prudence.create({
    type: 'router',
    name: 'myapp',
    routes: [{
        // Person resource
        paths: [ 'person/*' ],
        handler: prudence.import('person/resource.js')
    }, {
        // Static files
        handler: prudence.create({
            type: 'static',
            root: 'static/',
            indexes: [ 'index.html' ]
        })
    }, {
        // If nothing else matches
        handler: prudence.defaultNotFound
    }]
});
