
prudence.create({
    type: 'resource',
    name: 'person',
    facets: [{
        // Main facet
        name: 'main',
        paths: [ '{name}' ],
        representations: [{
            // A single JSON representation with a seperate describer and presenter
            hooks: prudence.hooks('main/json.js')
            //construct: prudence.hook('main/json.js', 'construct'),
            //describe: prudence.hook('main/json.js', 'describe'),
            //present: prudence.hook('main/json.js', 'present')
        }]
    }, {
        // Chores facet
        name: 'chores',
        paths: [ '{name}/chores' ],
        representations: [{
            // HTML representation using JST
            contentTypes: [ 'text/html' ],
            languages: [ 'en', 'he' ],
            present: prudence.hook('chores/html.jst', 'present')
        }, {
            // Default text representation
            present: prudence.hook('chores/text.js', 'present')
        }]
    }]
});
