
const library = require('../../library');

exports.prepare = library.prepare;

exports.present = function() {
    const o = {
        'name': this.variables.name,
        'app': this.variables.app,
        'resource': this.variables.resource,
        'facet': 'root'
    };

    const indent = (this.request.query.get('indent') === 'true') ? '  ' : '';
    this.transcribe(o, '', indent);
    this.log.info('resource1 root facet data for: ' + this.variables.name);
};
