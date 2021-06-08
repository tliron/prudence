Prudence: JavaScript Templates (JST)
====================================

The [tutorial](../README.md) covers basic usage, this page is for reference.

The tag delimiters are `<%` and `%>`. Characters that follow right after the opening delimiter
specify special tags, or "sugar".

The delimiters can be escaped by prefixing a backslash: `\<%` and `\%>`.

Most tags will will swallow the trailing newline character just after the tag's end delimiter.
This helps you avoid cluttering your output with empty lines. For example, this:

    No
    <% var x = 1; %>
    empty
    <% x += 1; %>
    lines!

will be output as this:

    No
    Empty
    Lines

To disable this feature use `/%>` as the closing delimiter. This:

    Empty
    <% var x = 1; /%>
    line!

will be output as this:

    Empty

    line!

The one exception is the "expression" tag, `<%=`, which does not swallow the trailing newline
because it's intended to be used within flows of text.

The following tags are built in. It is also possible to [extend](../platform/README.md#jst-sugar)
Prudence with additional tags.

### Cache duration: `<%* numeric_expr %>`

Equivalent to:

    context.cacheDuration = numeric_expr;

### Capture: `<%! name_expr %>` and `<%!!%>`

Captures the enclosed text into a context variable. Does *not* write it. Example:

    <%! 'greeting' %>
    <div>
        Hello, <%== 'name' %>!
    </div>
    <%!!%>

    The greeting is: <%== 'greeting' %>

See "Embed" for possible uses.

### Comment: `<%# anything %>`

The content of the tag is ignored. Can be useful for quickly disabling other tags
during development. Example:

    <%#
    This is
    ignored
    %>

Note that this does not even insert JavaScript comments, so though functionally
equivalent it is not quite identical to this:

    <%
    // This is
    // ignored
    %>

### Embed: `<%& filename_expr %>` or `<%& filename_expr, cachekey_expr %>`

This tag allows another representation to write itself into this one. Both the "construct"
and "present" hooks of the embedded representation are called, which means that if that
representation has been cached (with a "context.cacheDuration" > 0) then it may retrieve
from a cache. This can allow for powerful, fine-grained caching.

Note that the current context will be copied into the embedded context. This means that any
variable you set can be accessed in the embedded context. The opposite does not happen in
order to ensure that anything the embedded representation does to the context will not
interfere with the current one.

The long form allows you to set the "context.cacheKey" for the embedded representation
*before* embedding it. Note that if you do so then you should *not* change the cache key
in the embedded representation's "construct".

Both ".js" and ".jst" files are supported. Example:

    The menu is: <%& 'fragments/menu.jst', 'menu:' + context.variables.name %>

### Expression: `<%= expr %>`

Write a JavaScript expression. Example:

    Hello! Your name is <%= context.variables.name.toUpperCase() %>!

Equivalent to:

    context.writeString(String(expr));

### Insert: `<%+ filename_expr %>` or `<%+ filename_expr, renderer_expr %>`

Write the contents of a file, optionally rendering it first. Example:

    <%+ '../docs/README.md', 'markdown' %>

The short form is equivalent to:

    context.writeString(prudence.loadString(filename_expr));

The long form is equivalent to:

    context.writeString(prudence.render(prudence.loadString(filename_expr), renderer_expr));

### Render: `<%^ renderer_expr %>` and `<%^^%>`

Renders the enclosed text before writing it. Example:

    <%^ 'markdown' %>
    This is Markdown
    ================

    Hello, <%== 'name' %>!

    It is a *markup* language for generating HTML.
    <%^^%>

Note that any other JST tags inside the enclosed text are processed as usual.

### Scriptlet: `<%`

Insert any JavaScript code. Example:

    <% for (var i = 0; i < 10; i++) { %>
        <p>Number <%= i %></p>
    <% } %>

### Signature: [`<%$%>` or `<%$ bool_expr %>`] and `<%$$%>`

Automatically creates a "context.signature" based on a hash of the enclosed text,
which is otherwise written as is. By default it will be a strong signature. The
longer form, with the "bool_expr", will set the signature to weak if the "bool_expr"
is true. Example of a weak signature:

    <%$ true %>
    Hello, <%== 'name' %>!
    <%$$%>

Note that any other JST tags inside the enclosed text are processed as usual.

### Variable: `<%== name_expr %>`

Write a variable. Example:

    Hello! Your name is <%== 'name' %>!

Equivalent to:

    <%= context.variables[name_expr] %>
