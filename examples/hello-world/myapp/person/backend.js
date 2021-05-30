
const db = {};
const lock = prudence.mutex();

// Gets the signature *instead of the data*
// Let's pretend that it is less expensive than "getPerson"
exports.getSignature = function(name) {
    prudence.log.info('getSignature');
    return prudence.hash(name);
};

exports.getCachePrefix = function(name) {
    prudence.log.info('getCachePrefix');
    return 'person.' + name;
};

// Gets the data
// Let's pretend that it is from a database
// (And thus it's the most expensive part of any request)
exports.getPerson = function(name) {
    prudence.log.info('getPerson');
    lock.lock();
    try {
        var data = db[name];
        if (!data) {
            data = {
                name: name,
                chores: [ 'sleeping' ]
            };
            db[name] = data;
        }
        return prudence.deepCopy(data);
    } finally {
        lock.unlock();
    }
};

exports.deletePerson = function(name) {
    prudence.log.info('deletePerson');
    lock.lock();
    try {
        delete db[name];
    } finally {
        lock.unlock();
    }
};

exports.setChores = function(name, chores) {
    prudence.log.info('setChores');
    lock.lock();
    try {
        db[name] = {
            name: name,
            chores: chores
        };
    } finally {
        lock.unlock();
    }
};
