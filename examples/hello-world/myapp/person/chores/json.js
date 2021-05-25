
const backend = prudence.require('../backend.js');

function present(context) {
    context.log.info('present');
    context.contentType = 'application/json';
    prudence.encode(backend.getPerson(context.variables.name).chores, 'json', '  ', context);
}

function erase(context) {
    context.log.info('erase');
    backend.setChores(context.variables.name, []);
    present(context);
}

function change(context) {
    context.log.info('change');
    const chores = prudence.decode(context.request(), 'json');
    backend.setChores(context.variables.name, chores);
    present(context);
}
