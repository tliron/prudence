
prudence.create({
    type: 'resource',
    name: 'person',
    facets: [{
        // Main facet
        name: 'main',
        paths: '{name}',
        representations: { // can also be a list
            // A single JSON representation with all functions in a separate file
            functions: prudence.require('main/json.js'),
            // We can also set functions individually, like so:
            // runtime: runtime,
            // construct: prudence.require('main/json.js').construct,
            // present: function(context) { context.write('example'); }
        }
    }, {
        // Chores facet
        name: 'chores',
        paths: '{name}/chores',
        representations: [{
            // HTML representation using JST
            contentTypes: 'text/html', // can also be a list
            languages: [ 'en', 'he' ], // can be a list or a single language
            functions: prudence.require('chores/html.jst')
        }, {
            // Default representation (JSON)
            functions: prudence.require('chores/json.js')
        }]
    }]
});
