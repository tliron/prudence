
const backend = prudence.require('../backend.js');

// The "construct" hook (optional) optimizes for server-side caching
// It should be a very fast function
// Here we can set cacheKey, cacheDuration, contentType, charSet, and language
// cacheKey defaults to the request path
// contentType defaults to that of the accepted representation
function construct(context) {
    context.log.info('construct');
    context.cacheKey = 'person|' + context.variables.name;
    context.cacheDuration = 5;
    context.contentType = 'application/json';
}

// The "describe" hook (optional) optimizes for client-side caching
// It should be a fast function
// We can set the signature or lastModified here
// Safely assume that "construct" has already been called (if it exists)
function describe(context) {
    context.log.info('describe');
    context.signature = backend.getSignature(context.variables.name);
}

// The "present" hook (required) generates the representation
// Safely assume that "describe" has already been called (if it exists)
function present(context) {
    context.log.info('present');
    prudence.encode(backend.getData(context.variables.name), 'json', '  ', context);
    // The above is equivalent to this:
    //context.write(JSON.stringify(backend.getData(context), null, '  ')+'\n');
}
