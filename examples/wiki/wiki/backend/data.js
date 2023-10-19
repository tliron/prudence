
const PAGE_PREFIX = 'page:';

util.once('wiki.badger', function() {
    env.variables.wiki = env.variables.wiki || {};
    env.variables.wiki.db = badger.open(os.joinFilePath(__dirname, 'db'));
});

exports.init = function(db) {
    env.variables.wiki.db.update(function(transaction) {
        for (let path in db) {
            const key = PAGE_PREFIX + path;
            try {
                transaction.get(key);
            } catch(_) {
                transaction.set(key, db[path]);
            }
        }
    });
};

exports.get = function(path) {
    var entry;
    env.variables.wiki.db.view(function(transaction) {
        try {
            transaction.get(PAGE_PREFIX + path).value(function(value) {
                entry = value;
            });
        } catch (_) {}
    });
    return entry;
};

exports.set = function(path, entry) {
    env.variables.wiki.db.update(function(transaction) {
        transaction.set(PAGE_PREFIX + path, entry);
    });
};

exports.list = function() {
    const entries = {};
    env.variables.wiki.db.view(function(transaction) {
        transaction.iterate(function(item) {
            const key = item.key().slice(PAGE_PREFIX.length);
            item.value(function(value) {
                //console.log(transcribe.stringify(value, 'json'));
                entries[key] = value.title;
            });
        }, {
            prefix: PAGE_PREFIX
        });
    });
    return entries;
};

