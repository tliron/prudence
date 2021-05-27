

const db = {};
const lock = prudence.mutex();

// Gets the signature *instead of the data*
// Let's pretend that it is less expensive than "getPerson"
function getSignature(name) {
    prudence.log.info('getSignature');
    return prudence.hash(name);
}

function getCachePrefix(name) {
    prudence.log.info('getCachePrefix');
    return 'person|' + name;
}

// Gets the data
// Let's pretend that it is from a database
// (And thus it's the most expensive part of any request)
function getPerson(name) {
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
        data = prudence.deepCopy(data);
        return data
    } finally {
        lock.unlock();
    }
}

function deletePerson(name) {
    prudence.log.info('deletePerson');
    lock.lock();
    try {
        delete db[name];
    } finally {
        lock.unlock();
    }
}

function setChores(name, chores) {
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
}
