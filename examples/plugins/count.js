
function present() {
    util.once('badger', function() {
        // Open a Badger database (via our API plugin)
        env.variables.db = badger.open(os.joinFilePath(__dirname, 'db'));
    });

    let counter = 0;

    // See: https://pkg.go.dev/github.com/dgraph-io/badger/v4
    env.variables.db.update(function(transaction) {
        try {
            transaction.get('counter').value(function(value) {
                counter = value;
            });
        } catch (_) {}

        transaction.set('counter', counter + 1);
    });

    this.transcribe({counter: counter});
}

exports.handler = new prudence.Router({
    routes: [{
        paths: 'count',
        handler: new prudence.Resource({
            facets: {
                representations: {
                    contentTypes: prudence.dataContentTypes,
                    present: present
                }
            }
        }),
    }, {
        paths: 'echo',
        handler: new prudence.Echo({ // our type plugin
            message: 'Hello from "Echo"!'
        })
    }, {
        handler: prudence.notFound
    }]
});