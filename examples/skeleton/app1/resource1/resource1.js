
const library = require('../library');

exports.handler = new prudence.Resource({
    name: 'resource1',
    facets: [{
        name: 'root',
        paths: [ '{name}', '{name}/' ],
        variables: {
            resource: 'resource1'
        },
        representations: [{
            name: 'data',
            contentTypes: prudence.dataContentTypes,
            languages: 'en',
            hooks: require('./root-facet/data')
        }, {
            name: 'html',
            contentTypes: 'text/html',
            languages: 'en',
            redirectTrailingSlash: true,
            prepare: library.prepare,
            present: require('./root-facet/html.jst').present
        }]
    }, {
        name: 'sub1',
        paths: [ '{name}/sub1', '{name}/sub1/' ],
        representations: [{
            name: 'data',
            contentTypes: prudence.dataContentTypes,
            languages: 'en',
            hooks: require('./sub-facet1/data')
        }, {
            name: 'html',
            contentTypes: 'text/html',
            languages: 'en',
            redirectTrailingSlash: true,
            prepare: library.prepare,
            present: require('./sub-facet1/html.jst').present
        }]
    }]
});
