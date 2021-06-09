
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{  // can be a list or a single route
        handler: function(context) {
            if (context.path === 'magic') {
                context.writeString('Magic!\n')
                context.response.addCookie({name: 'mycookie', value: 'myvalue'});
                return true; // handled
            } else {
                return false; // not handled
            }
        }
    }, {
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
