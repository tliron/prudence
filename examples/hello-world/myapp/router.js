
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{  // can be a list or a single route
        handler: function(context) {
            if (context.request.path === 'magic') {
                let value = 'novalue'

                // Receive a cookie
                let cookie = context.request.getCookie('mycookie');
                if (cookie) {
                    value = cookie.value;
                    context.log.infof('mycookie value = %s', value);
                }

                // Send a cookie
                context.write('Magic!\n')
                context.response.addCookie({name: 'mycookie', value: value + '-magic'});
                return true; // handled
            } else {
                return false; // not handled
            }
        }
    }, {
        // Person resource
        paths: 'person/*', // can also be a list
        handler: bind('./person/resource', 'handler')
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
