
function represent(context) {
    context.log.info('json!');
    context.log.infof('%T', context);
    context.write(JSON.stringify({name: context.variables.name}));
}
