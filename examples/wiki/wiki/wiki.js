
exports.handler = new prudence.Router({
    name: 'wiki',
    routes: [{
        paths: 'static/*',
        handler: new prudence.Static({
            root: 'static/',
            indexes: 'index.html'
        })
    }, {
        handler: require('./page/page').handler
    }]
});
