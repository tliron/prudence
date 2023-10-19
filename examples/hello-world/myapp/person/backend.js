
// "prudence.once" makes sure that the function is only ever called
// a single time, which is useful for global initialization
prudence.once('backend', function() {
    jst.log.info('initializing backend');
    prudence.globals.db = {};
    prudence.globals.dbLock = prudence.mutex();
});

// Gets the signature *instead of the data*
// Let's pretend that it is less expensive than "getPerson"
exports.getSignature = function(name) {
    jst.log.info('getSignature');
    return prudence.hash(name);
};

// Gets the data
// Let's pretend that it is from a database
// (And thus it's the most expensive part of any request)
exports.getPerson = function(name) {
    jst.log.info('getPerson');
    prudence.globals.dbLock.lock();
    try {
        let data = prudence.globals.db[name];
        if (!data) {
            data = {
                name: name,
                chores: [ 'sleeping' ]
            };
            prudence.globals.db[name] = data;
        }
        return prudence.deepCopy(data);
    } finally {
        prudence.globals.dbLock.unlock();
    }
};

exports.deletePerson = function(name) {
    jst.log.info('deletePerson');
    prudence.globals.dbLock.lock();
    try {
        delete prudence.globals.db[name];
    } finally {
        prudence.globals.dbLock.unlock();
    }
};

exports.setChores = function(name, chores) {
    jst.log.info('setChores');
    prudence.globals.dbLock.lock();
    try {
        prudence.globals.db[name] = {
            name: name,
            chores: chores
        };
    } finally {
        prudence.globals.dbLock.unlock();
    }
};
