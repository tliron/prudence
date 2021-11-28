
exports.getCachePrefix = function(context) {
    return 'person.' + context.variables.name;
};

exports.addCacheGroup = function(context) {
    context.cacheGroups.push(exports.getCachePrefix(context));
};
