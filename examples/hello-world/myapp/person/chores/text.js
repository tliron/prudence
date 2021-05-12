
function present(context) {
    context.contentType = 'text/plain';
    context.write('Chores for: ' + context.variables.name + '\n');
}
