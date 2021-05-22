
// Gets the data
// Let's pretend that it is from a database
// (And thus it's the most expensive part of any request)
function getData(name) {
    prudence.log.info('getData');
    return {
        name: name,
        chores: [ 'cleaning', 'shopping', 'cooking' ]
    };
}

// Gets the signature *instead of the data*
// Let's pretend that it is less expensive than "getData"
function getSignature(name) {
    prudence.log.info('getSignature');
    return prudence.hash(name);
}
