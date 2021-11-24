
exports.handler = new prudence.Resource({
    name: 'person',
    facets: [{
        // Main facet
        name: 'main',
        paths: '{name}',
        representations: { // can also be a list
            // A single JSON representation with all functions in a separate file
            functions: bind('./main/json')
            // We can also set functions individually, like so:
            // construct: bind('./main/json', 'construct'),
            // present: function(context) { context.write('example'); }
        }
    }, {
        // Chores facet
        name: 'chores',
        paths: '{name}/chores',
        representations: [{
            // Default representation (JSON)
            functions: bind('./chores/json', 'Chores')
        }, {
            // HTML representation using JST
            contentTypes: 'text/html', // can also be a list
            languages: [ 'en', 'he' ], // can be a list or a single language
            functions: bind('./chores/html.jst'),
            construct: function(context) {
                const backend = require('./backend');
                const cachePrefix = backend.getCachePrefix(context.variables.name);
                context.cacheGroups.push(cachePrefix);
            }
        }]
    }]
});
