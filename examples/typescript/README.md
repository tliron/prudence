Prudence: TypeScript Example
============================

[TypeScript](https://www.typescriptlang.org/) is JavaScript on steroids. It's fully compatible
with JavaScript but is a vastly more powerful and more elegant language. Its best feature is that
it's typed so that many errors can be avoided before runtime. Also, typed languages are more
pleasurable work with in IDEs, with support for code completion, refactoring, and more.

Prudence comes with a [base TypeScript project](../../assets/typescript/prudence/) that declares
all the built-in APIs and is ready to use. All you need to do is to extend it in your project's
`tsconfig.json`, like we do in this example:

    {
        "extends": "../../assets/typescript/prudence/tsconfig.json",
        "include": [ "**/*" ]
    }


How to Run
----------

Like most other TypeScript environments Prudence doesn't execute TypeScript code directly. Instead,
it expects TypeScript to be transpiled into JavaScript.

Make sure the transpiler, [`tsc`](https://www.typescriptlang.org/docs/handbook/compiler-options.html),
is installed. It may very well be included in your operating system's repository. For example, in the
Fedora world:

    sudo dnf install nodejs-typescript

Or in the Debian world:

    sudo apt install node-typescript

You can then run `tsc` to transpile your `.ts` files into `.js` files and even use `--watch` so that
changes will be picked up and re-transpiled. When combined with Prudence's `--watch=true` (the default)
this will allow live editing of TypeScript. For example:

    tsc --watch --project examples/typescript

And then run Prudence in another terminal:

    prudence run examples/typescript/start.js

If you edit `start.ts` then `start.js` will also be updated and Prudence will restart the servers.


A Note on Modules
-----------------

TypeScript's "import" statement refers to other `.ts` files, but this is transpiled into a JavaScript
"require" with the same name, which will actually require the corresponding `.js` file. Prudence
itself is entirely unaware of the existence of `.ts` files.

On the other hand, TypeScript is unaware that "bind" is a special kind of "import". The bind will work,
but TypeScript will not actually check that the bound file implements the right hooks.
