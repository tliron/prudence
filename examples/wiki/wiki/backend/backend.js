
const data = require('./data');

util.once('wiki.init', function() {
    const initial = ard.decode(env.loadBytes('./initial.yaml'), 'yaml')
    //console.log(transcribe.stringify(initial, 'json'));
    data.init(initial);
});

class Page {
    constructor(path, entry) {
        this.path = path;
        if (entry) {
            this.title = entry.title;
            this.content = entry.content;
        } else {
            this.title = '[title]';
            this.content = '[content]';
        }
    }

    save() {
        data.set(this.path, {
            title: this.title,
            content: this.content
        });
    }
}

class Pages {
    constructor(context) {
        this.context = context;
    }

    get urlRoot() {
        return this.context.variables.wiki.root;
    }

    get currentPath() {
        return this.context.request.path.slice(0, -1);
    }

    get index() {
        return data.list();
    }

    exists(path) {
        if (!path)
            path = this.currentPath;
        const entry = data.get(path);
        if (entry)
            return true;
        else
            return false;
    }

    get(path) {
        if (!path)
            path = this.currentPath;
        const entry = data.get(path);
        if (entry)
            return new Page(path, entry)
        else
            return null;
    }

    new(path) {
        if (!path)
            path = this.currentPath;
        return new Page(path);
    }
    
    viewUrl(path) {
        if (!path)
            path = this.currentPath;
        return util.url({path: this.urlRoot + path + '/'});
    }
    
    editUrl(path) {
        if (!path)
            path = this.currentPath;
       return util.url({path: this.urlRoot + path + '/', query: {edit: 'true'}})    
    }

    url(path) {
        return util.url({path: this.urlRoot + path});
    }
}

exports.Pages = Pages;
