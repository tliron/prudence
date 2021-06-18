
export const resource = new prudence.Resource({
    facets: {
        representations: {
            functions: bind('./main')
        }
    }
});
