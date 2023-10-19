
exports.handler = new prudence.Router({
    routes: [{
        // The "*" wildcard captures the path segment into the path for the handler
        // Doubling the trailing "/" forces a "/" via a 301 redirect ("app1" to "app1/")
        paths: 'app1//*',
        handler: require('./app1/app1').handler
    }, {
        // Redirect legacy URLs
        paths: 'oldapp1//*',
        handler: function(context) {
            context.redirect(util.url({
                path: '/app1/' + context.request.path,
                query: context.request.query
            }), 301); // defaults to 302
            return true;
        }
    }, {
        // Simple Not Found (404) response in text/plain
        handler: prudence.notFound
    }]
});
