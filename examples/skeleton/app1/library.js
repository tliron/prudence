
exports.prepare = function() {
    this.cacheKey = this.variables.app + ';' + this.variables.resource + ';' + this.request.path;
    this.cacheDuration = 5;
};
