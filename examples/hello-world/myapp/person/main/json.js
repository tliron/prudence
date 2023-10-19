
const backend = require('../backend');
const common = require('../common');

// The "construct" hook (optional) optimizes for server-side caching
// It should be a fast function
// Here we can set "cacheKey", "cacheGroups", "contentType", "charSet", and "language"
// "cacheKey" defaults to the request path
// "contentType" defaults to that of the accepted representation
exports.construct = function(context) {
    context.log.info('construct');
    common.addCacheGroup(context);
    context.cacheKey = common.getCachePrefix(context) + '.main';
    context.response.contentType = 'application/json';
};

// The "describe" hook (optional) optimizes for client-side caching
// It should be a fast function
// We can set "signature" or "timestamp" here
// Safely assume that "construct" has already been called (if it exists)
exports.describe = function(context) {
    context.log.info('describe');
    context.response.signature = backend.getSignature(context.variables.name);
};

// The "present" hook (required) generates the representation
// Safely assume that "describe" has already been called (if it exists)
exports.present = function(context) {
    context.log.info('present');
    context.cacheDuration = 5;
    context.writeJson(backend.getPerson(context.variables.name), '  ');
};
