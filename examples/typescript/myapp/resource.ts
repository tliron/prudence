
declare const prudence: any;
declare const bind: any;

export const resource = new prudence.Resource({
    facets: {
        representations: {
            functions: bind('./main.ts')
        }
    }
});
