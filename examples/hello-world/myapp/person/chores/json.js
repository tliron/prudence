
const backend = prudence.require('../backend.js');

function present(context) {
    context.contentType = 'application/json';
    prudence.encode(backend.getData(context.variables.name).chores, 'json', '  ', context);
}
