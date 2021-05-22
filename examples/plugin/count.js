
// A Badger database
// See: https://pkg.go.dev/github.com/dgraph-io/badger/v3#section-documentation
const db = myplugin.badger(prudence.here() + '/db');

function present(context) {
    var counter;

    db.update(function(txn) {
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

    context.write('Counter is ' + counter + '\n');
}
