
prudence.once('db', function() {
    // A Badger database
    // See: https://pkg.go.dev/github.com/dgraph-io/badger/v3#section-documentation
    prudence.globals.db = myplugin.badger(prudence.joinFilePath(__dirname, 'db'));
});

exports.present = function(context) {
    let counter;

    prudence.globals.db.update(function(txn) {
        try {
            txn.get('counter').value(function(value) {
                counter = parseInt(prudence.bytesToString(value));
                return null;
            });
        } catch (e) {
            counter = 0;
        }

        txn.set('counter', prudence.stringToBytes(counter + 1));

        return null;
    });

    context.response.contentType = 'application/json'
    context.write('{"counter": ' + counter + '}\n');
};

