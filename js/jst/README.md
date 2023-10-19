Prudence: JST
=============

### Embed: `<%& filename_expr %>`

This tag allows another representation to write itself into this one. Only the "present"
hook of the embedded representation is called.

Note that the current context will be copied into the embedded context. This means that any
"this.variable" you set can be accessed in the embedded context. The opposite does not happen
in order to help ensure that anything the embedded representation does to the context will not
interfere with the current one.

### Cached Embed: `<%&& filename_expr, cacheduration_expr, cachekey_expr %>` or `<%&& filename_expr, cacheduration_expr, cachekey_expr, cahegroups_expr %>`

However, the embedded representation can be
cached if it sets "context.cacheDuration" > 0 in its "present". This is suitable for `.jst`
files, which only have a "present" hook.

The long forms allow you to set the "context.cacheKey" and even "context.cacheGroups" for the
embedded representation *before* embedding it, acting as a simple version of "construct".

Example of longest form:

    The menu is: <%& 'fragments/menu.jst', 5, 'menu:' + context.variables.name, ['person'] %>

### Signature: [`<%$%>` or `<%$ bool_expr %>`] and `<%$$%>`

Automatically creates a "context.signature" based on a hash of the enclosed text,
which is otherwise written as is. By default it will be a strong signature. The
longer form, with the "bool_expr", will set the signature to weak if the "bool_expr"
is true. Example of a weak signature:

    <%$ true %>
    Hello, <%== 'name' %>!
    <%$$%>

Note that any other JST tags inside the enclosed text are processed as usual.

