
prudence.create({
    type: 'resource',
    name: 'person',
    facets: [{
        name: 'main',
        paths: [ '{name}' ],
        representations: [{
            // A single JSON representation with a seperate describer and presenter
            describer: prudence.hook('main/json.js', 'describe'),
            presenter: prudence.hook('main/json.js', 'present')
        }]
    }, {
        name: 'chores',
        paths: [ '{name}/chores' ],
        representations: [{
            // HTML representation using JST
            contentTypes: [ 'text/html' ],
            presenter: prudence.hook('chores/html.jst', 'present')
        }, {
            // Default text representation
            presenter: prudence.hook('chores/text.js', 'present')
        }]
    }]
});
