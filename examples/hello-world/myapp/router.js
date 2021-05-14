
prudence.create({
    type: 'router',
    name: 'myapp',
    routes: [{ // can be a list or a single route
        // Person resource
        paths: 'person/*', // can also be a list
        handler: prudence.import('person/resource.js')
    }, {
        // Static files
        handler: prudence.create({
            type: 'static',
            root: 'static/',
            indexes: 'index.html' // can also be a list
        })
    }, {
        // If nothing else matches
        handler: prudence.defaultNotFound
    }]
});
