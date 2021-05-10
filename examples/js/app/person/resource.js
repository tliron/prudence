
prudence.create({
    type: 'resource',
    name: 'person',
    facets: [{
        name: 'main',
        paths: [ '{name}' ],
        representations: [{
            // JSON
            contentTypes: [ 'application/json' ],
            representer: prudence.hook('main/json.js', 'represent')
        }, {
            // HTML
            contentTypes: [ 'text/html' ],
            representer: prudence.hook('main/html.jst', 'represent') // JST
        }, {
            // Default representation
            representer: prudence.hook('main/text.js', 'represent')
        }]
    }, {
        // Age facet
        name: 'age',
        paths: [ '{name}/age' ],
        representations: [{
            representer: prudence.hook('age/html.js', 'represent')
        }]
    }]
});
