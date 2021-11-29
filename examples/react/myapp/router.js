
exports.handler = new prudence.Router({
    name: 'myapp',
    routes: [{
        handler: new prudence.Static({
            root: 'static/',
            indexes: 'index.html'
        })
    }, {
        handler: prudence.defaultNotFound
    }]
});
