
// Gets the data
// Let's pretend that is from a database
// And thus it's the most expensive part of any request
function getData(context) {
    context.log.info('getData');
    return {
        name: context.variables.name
    };
}

// Gets the signature *instead of the data*
// Let's pretend that it is less expensive than "getData"
function getSignature(context) {
    context.log.info('getSignature');
    return prudence.hash(context.variables.name);
}

// OPTIONAL HOOK:
// "construct" optimizes for server-side caching
// It should be a very fast function
// Here we can set cacheKey, cacheDuration, contentType, charSet, and language
// cacheKey defaults to the request path
// contentType defaults to that of the accepted representation
function construct(context) {
    context.log.info('construct');
    context.cacheKey += '-test'
    context.cacheDuration = 5;
    context.contentType = 'application/json';
}

// OPTIONAL HOOK:
// "describe" optimizes for client-side caching
// It should be a fast function
// We can set the ETag or lastModified signatures here
// Safely assume that "construct" has already been called (if it exists)
//
// For this to be a meaningful optimization it must be *less expensive* then "present"
// This can be challenging and even impossible to achieve in some cases. Example:
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
function describe(context) {
    context.log.info('describe');
    context.ETag = getSignature(context);
}

// REQUIRED HOOK:
// "present" generates the representation
// Safely assume that "describe" has already been called (if it exists)
function present(context) {
    context.log.info('present');
    prudence.encode(getData(context), 'json', '  ', context);
    // The above is equivalent to this:
    //context.write(JSON.stringify(getData(context), null, '  ')+'\n');
}
