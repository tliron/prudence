
const backend = prudence.require('../backend.js');

function present(context) {
    context.contentType = 'application/json';
    prudence.encode(backend.getPerson(context.variables.name).chores, 'json', '  ', context);
}
