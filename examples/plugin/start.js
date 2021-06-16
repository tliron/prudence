
// The "myplugin" global is from our plugin
myplugin.print('Hello from our API!');

const resource = new prudence.Resource({
    facets: {
        representations: {
            functions: bind('./count')
        }
    }
});

const router = new prudence.Router({
    routes: [{
        paths: 'count',
        handler: resource
    }, {
        handler: new prudence.Echo({ // from our plugin
            message: 'Hello from "echo"!'
        })
    }]
});

prudence.start(new prudence.Server({
    address: 'localhost:8080',
    handler: router
}));
