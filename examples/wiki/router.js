
exports.handler = new prudence.Router({
    routes: [{
        paths: 'wiki//*',
        variables: {
            wiki: {
                name: 'Wiki',
                root: '/wiki/'
            }
        },
        handler: require('./wiki/wiki').handler
    }, {
        handler: prudence.notFound
    }]
});
