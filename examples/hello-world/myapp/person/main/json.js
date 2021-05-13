
// Gets the data
// This could potentially be from a database,
// and would thus be the most expensive part of any request
function getData(context) {
    context.log.info('getData');
    return {
        name: context.variables.name
    };
}

// Gets the signature *instead of the data*

// For this to be a meaningful optimization it must be *less expensive* then getData
// This can be challenging and even impossible in some cases. Example:
//
// * Let's say that getData involves a complex SQL join of many tables
// * We'll create another table that just has two columns, the ID and the signature
// * Thus it would be much cheaper to access this table here in getSignature
// * The signature *does not* have to reflect the actual data, it just has to
//   be different every time the data changes
// * Thus it can be as simple as a random value (ETag) or a last-modified timestamp
// * But the big challenge is that the backend would need to make sure to delete
//   these signatures if *any of the dependent rows change*
// * If it's difficult to catch such changes it may be possible to err on the side of
//   deleting signatures even when unsure, for example if a certain table changes
//   then all signatures of a certain known class can be invalidated
// * Of course if signatures are invalidated too frequently this optimization will
//   not be meaningful
function getSignature(context) {
    context.log.info('getHash');
    return prudence.hash(context.variables.name);
}

// Describes the representation
// We should *only* set the ETag or LastModified here, not write content
function describe(context) {
    context.log.info('describe');
    context.ETag = getSignature(context);
}

// Presents the representation
// Often "present" will also call "describe", because the presentation should also include the description
function present(context) {
    context.log.info('present');
    describe(context);
    context.contentType = 'application/json';
    context.write(JSON.stringify(getData(context), null, '  ')+'\n');
    //prudence.encode(context.scratch.data.content, 'json', '  ', context);
}
