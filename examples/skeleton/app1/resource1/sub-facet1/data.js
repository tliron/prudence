
const library = require('../../library');

exports.prepare = library.prepare;

exports.present = function() {
    const o = {
        'name': this.variables.name,
        'app': this.variables.app,
        'resource': this.variables.resource,
        'facet': 'sub1'
    };

    const indent = (this.request.query.get('indent') === 'true') ? '  ' : '';
    this.transcribe(o, '', indent);
    this.log.info('resource1 sub1 facet data for: ' + this.variables.name);
};
