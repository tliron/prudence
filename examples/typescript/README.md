Prudence: TypeScript Example
============================

If you prefer [TypeScript](https://www.typescriptlang.org/) then Prudence will automatically
transpile it to JavaScript for you. To do this Prudence needs the
[`tsc` command](https://www.typescriptlang.org/docs/handbook/compiler-options.html) installed.

It may already be availabe for your operating system. For example, in the Fedora world:

    sudo dnf install nodejs-typescript

In the Debian world:

    sudo apt install node-typescript

You can then refer to TypeScript files as you would to JavaScript files:

    prudence run examples/typescript/start.ts


Features and Limitations
------------------------

Everything in TypeScript needs to be declared. For now we do not have declarations for the
Prudence APIs, but you can get away with declaring them as "any":

    declare const prudence: any;
    declare const bind: any;

Prudence can only transpile TypeScript if it's in the local filesystem, not on URLs. Transpiling
will place a JavaScript file in the same directory as your TypeScript file with just the extension
changed from `.ts` to `.js`. For example, `myapp/main.ts` will produce `myapp/main.js`.

In TypeScript you will "import" and "export" instead of CommonJS's "require" an "exports", e.g.:

    import {resource} from './myapp/resource';

Note that "require" lets you specify a complete path, including an extension, however "import"
is stricter and does not allow extensions. TypeScript will be using the `.ts` file, but Prudence
will be using the `.js` file.

A consequence of this is that when using `--watch=true` (the default) Prudence will restart on changes
to `.ts` files *only if* you referred to them *directly*. In the case of this example it would be
`start.ts`, which we use in the `prudence run` command, and `lib/main.ts`, which we use in the call to
"bind" in `resource.ts`. Thus if you change `resource.ts` Prudence will not restart because it never
refers to it directly. You can get around this limitation via a quick `touch` to any file Prudence does
knows about, e.g.:

    touch examples/typescript/start.ts
