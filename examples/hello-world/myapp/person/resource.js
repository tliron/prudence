
prudence.create({
    type: 'resource',
    name: 'person',
    facets: [{
        // Main facet
        name: 'main',
        paths: '{name}',
        representations: { // can also be a list
            // A single JSON representation with a seperate describer and presenter
            hooks: prudence.hooks('main/json.js')
            // Can also set hooks individually, like so:
            //construct: prudence.hook('main/json.js', 'construct'),
            //describe: prudence.hook('main/json.js', 'describe'),
            //present: prudence.hook('main/json.js', 'present')
        }
    }, {
        // Chores facet
        name: 'chores',
        paths: '{name}/chores',
        representations: [{
            // HTML representation using JST
            contentTypes: 'text/html', // can also be a list
            languages: [ 'en', 'he' ], // can be a list or a single language
            hooks: prudence.hooks('chores/html.jst')
        }, {
            // Default text representation
            hooks: prudence.hooks('chores/text.js')
        }]
    }]
});
