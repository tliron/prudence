

// Gets the data (this could potentially be from a database)
// Ideally the data source contains some kind of signature or hash identifying the content
// In this case we will calculate a hash ourselves
// We'll store it in the "scratch" area so that the other functions can use it
function get(context) {
    context.log.info('get');
    const content = {
        name: context.variables.name
    };
    context.scratch.data = {
        content: content,
        hash: prudence.hash(content)
    };
}

// Describes the representation, which means *only* setting the ETag or LastModified
function describe(context) {
    get(context);
    context.log.info('describe');
    context.ETag = context.scratch.data.hash;
}

// Presents the representation
// Often "present" will also call "describe", because the presentation should also include the description
function present(context) {
    describe(context);
    context.log.info('present');
    context.contentType = 'application/json';
    context.write(JSON.stringify(context.scratch.data.content, null, '  ')+'\n');
    //prudence.encode(context.scratch.data.content, 'json', '  ', context);
}
