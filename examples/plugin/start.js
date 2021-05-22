
// The "myplugin" global is from our plugin
myplugin.print('Hello from our API!');

const resource = prudence.create({
    type: 'resource',
    facets: {
        representations: {
            functions: prudence.require('count.js')
        }
    }
});

const router = prudence.create({
    type: 'router',
    routes: [{
        paths: 'count',
        handler: resource
    }, {
        handler: prudence.create({
            type: 'myplugin.echo', // from our plugin
            message: 'Hello from "echo"!'
        })
    }]
});

prudence.start(prudence.create({
    type: 'server',
    address: 'localhost:8080',
    handler: router
}));
