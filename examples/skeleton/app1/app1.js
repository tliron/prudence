
exports.handler = new prudence.Router({
    name: 'app1',
    routes: [{
        // Resource 1
        paths: 'resource1/*',
        variables: {
            app: 'app1'
        },
        handler: require('./resource1/resource1').handler
    }, {
        // Static files
        handler: new prudence.Static({
            root: 'static/',
            indexes: 'index.html',
            presentDirectories: true
        })
    }]
});
