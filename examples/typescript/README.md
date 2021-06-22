Prudence: TypeScript Example
============================

[TypeScript](https://www.typescriptlang.org/) is JavaScript on steroids. It's fully compatible
with JavaScript but is a vastly more powerful and more elegant language. Its best feature is that
it's typed so that many errors can be avoided before runtime. Also, typed languages are more
pleasurable work with in IDEs, with support for code completion, refactoring, and more.

Prudence comes with a
[base TypeScript project](https://github.com/tliron/prudence/tree/main/assets/typescript/prudence/)
that declares all the built-in APIs and is ready to use. All you need to do is to extend it in your
project's `tsconfig.json`, like we do in this example:

    {
        "extends": "../../assets/typescript/prudence/tsconfig.json",
        "include": [ "**/*" ]
    }

Also see the [API documentation](https://prudence.threecrickets.com/assets/typescript/prudence/docs/).


How to Run
----------

Like most other TypeScript environments Prudence doesn't execute TypeScript code directly. Instead,
it transpiles TypeScript into JavaScript.

Make sure the transpiler, [`tsc`](https://www.typescriptlang.org/docs/handbook/compiler-options.html),
is installed. It may very well be included in your operating system's repository. For example, in the
Fedora world:

    sudo dnf install nodejs-typescript

Or in the Debian world:

    sudo apt install node-typescript

Prudence can now run `tsc` for you, including re-running it automatically when any `.ts` files are
changed (when `--watch=true`, the default). To enable TypeScript support use the `--typescript` flag
to point to the directory where your `tsconfig.json` is:

    prudence run examples/typescript/start.js --typescript=examples/typescript

Note that we are running the `.js` file. (If it doesn't exist yet, it will be created by transpiler
from `start.ts`.)

If you edit `start.ts` then `start.js` will also be updated and Prudence will restart the servers.


A Note on Modules
-----------------

TypeScript's "import" statement refers to other `.ts` files, but this is transpiled into a JavaScript
"require" with the same name, which will actually require the corresponding `.js` file. Prudence
itself is entirely unaware of the existence of `.ts` files.

On the other hand, TypeScript is unaware that "bind" is a special kind of "import". The bind will work,
but TypeScript will not actually check that the bound file implements the right hooks.
