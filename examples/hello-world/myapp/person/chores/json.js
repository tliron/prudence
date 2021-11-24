
const backend = require('../backend');

exports.Chores = {
    construct: function(context) {
        context.log.info('construct');
        const cachePrefix = backend.getCachePrefix(context.variables.name);
        context.cacheGroups.push(cachePrefix);
        context.cacheKey = cachePrefix + '.chores';
        context.response.contentType = 'application/json';
    },

    present: function(context) {
        context.log.info('present');
        context.writeJson(backend.getPerson(context.variables.name).chores, '  ');
    },

    describe: function(context) {
        context.log.info('describe');
        context.response.signature = backend.getSignature(context.variables.name);
    },

    erase: function(context) {
        context.log.info('erase');
        prudence.go(function() {
            backend.setChores(context.variables.name, []);
            prudence.invalidateCacheGroup(backend.getCachePrefix(context.variables.name));
        });
        context.done = true;
        context.async = true;
    },

    modify: function(context) {
        context.log.info('modify');
        const chores = prudence.decode(context.request.body, 'json');
        backend.setChores(context.variables.name, chores);
        prudence.invalidateCacheGroup(backend.getCachePrefix(context.variables.name));
        context.done = true;
        exports.Chores.present(context);
    },

    call: function(context) {
        context.log.info('call');
        exports.Chores.present(context);
    }
};
