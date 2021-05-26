
const backend = prudence.require('../backend.js');

function present(context) {
    context.log.info('present');
    context.contentType = 'application/json';
    prudence.encode(backend.getPerson(context.variables.name).chores, 'json', '  ', context);
}

function erase(context) {
    context.log.info('erase');
    prudence.go(function() {
        backend.setChores(context.variables.name, []);
    });
    context.done = true;
    context.async = true;
}

function change(context) {
    context.log.info('change');
    const chores = prudence.decode(context.request(), 'json');
    backend.setChores(context.variables.name, chores);
    context.done = true;
    present(context);
    // TODO:
    /// prudence.invalidateCacheTag(...)
}
