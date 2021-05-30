
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{ // can be a list or a single route
        // Person resource
        paths: 'person/*', // can also be a list
        handler: require('person/resource.js').handler
    }, {
        // Static files
        handler: new prudence.Static({
            root: 'static/',
            indexes: 'index.html' // can also be a list
        })
    }, {
        // If nothing else matches
        handler: prudence.defaultNotFound
    }]
});
