
const backend = require('../backend/backend');

const INDEX = 'index/';

const index = require('./index.jst').present;
const edit = require('./edit.jst').present;
const view = require('./view.jst').present;

function isIndex(context) {
    return context.request.path === INDEX;
}

function isEdit(context) {
    return context.request.query.get('edit') === 'true';
}

exports.handler = new prudence.Resource({
    name: 'page',
    facets: {
        representations: {
            contentTypes: 'text/html',
            languages: 'en',
            redirectTrailingSlash: true,
            prepare: function() {
                if (!isIndex(this) && !isEdit(this)) {
                    this.cacheKey = 'page;' + this.request.path;
                    this.cacheDuration = 5;
                }
            },
            present: function() {
                if (isIndex(this))
                    return index.call(this);
                else if (isEdit(this))
                    return edit.call(this);
                else
                    return view.call(this);
            },
            call: function() {
                if (isIndex(this)) {
                    this.response.status = 405; // method not allowed
                    this.write('Cannot edit index page');
                    this.end();
                    return;
                }

                const pages = new backend.Pages(this);

                const page = pages.new();
                page.title = this.request.direct.postFormValue('title');
                page.content = this.request.direct.postFormValue('content');
                page.save();

                this.cacheKey = 'page;' + this.request.path;
                this.deleteCachedRepresentation();
                this.redirect(pages.viewUrl());
            }
        }
    }
});
