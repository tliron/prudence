
declare function bind(id: string, name?: string): any;

declare interface Bytes {}

declare interface Writer {}

declare interface Logger {
    allowLevel(level: number): boolean;
    setMaxLevel(level: number): void;
    getMaxLevel(): number;

    newMessage(level: number, depth: number, ...keysAndValues: any[]): Message | null;

    log(level: number, depth: number, m: string, ...keysAndValues: any[]): void;
    logf(level: number, depth: number, f: string, ...args: any[]): void;

    critical(m: string, ...keysAndValues: any[]): void;
    criticalf(f: string, ...args: any[]): void;
    error(m: string, ...keysAndValues: any[]): void;
    errorf(f: string, ...args: any[]): void;
    warning(m: string, ...keysAndValues: any[]): void;
    warningf(f: string, ...args: any[]): void;
    notice(m: string, ...keysAndValues: any[]): void;
    noticef(f: string, ...args: any[]): void;
    info(m: string, ...keysAndValues: any[]): void;
    infof(f: string, ...args: any[]): void;
    debug(m: string, ...keysAndValues: any[]): void;
    debugf(f: string, ...args: any[]): void;
}

declare interface Message {
    set(k: string, v: any): Message;
    send(): void;
}

declare interface RestContext {
    variables: { [key: string]: any; };
    id: number;
    request: RestRequest;
    response: RestResponse;
    name: string;
    log: Logger;
    debug: boolean;
    done: boolean;
    created: boolean;
    async: boolean;
    cacheDuration: number;
    cacheKey: string;
    cacheGroups: string[];

    getVariable(...keys: any): any;
    write(content: any): void;
    transcribe(value: any, format: transcribe.Format, indent?: string): void;
    redirect(url: string, status?: number): void;
    redirectTrailingSlash(status?: number): void;
    internalServerError(): void;
    end(): void;
    clone(): RestContext;
}

declare interface RestRequest {
    host: string;
    port: number;
    path: string;
    header: Header;
    method: string;
    query: { [key: string]: string[]; };
    cookies: Cookie[];
    direct: any;

    body(): Bytes;
    bodyAsString(): string;
    getCookie(name: string): Cookie | null;
    clone(): RestRequest;
}

declare interface RestResponse {
    status: number;
    header: Header;
    cookies: Cookie[];
    contentType: string;
    charSet: string;
    language: string;
    signature: string;
    weakSignature: boolean;
    timestamp: number;
    bypass: boolean;
    direct: any;

    reset(): void;
    addCookie(config: CookieConfig): void;
    clone(): RestResponse;
}

declare interface Header {
    set(key: string, value: string): void;
    add(key: string, value: string): void;
    get(key: string): string;
    values(key: string): string[];
    del(key: string): void;
    clone(): Header;
}

declare type Cookie = {
    name: string;
    value: string;
    path: string;
    domain: string;
    expires: any;
    maxAge: number;
    secure: boolean;
    httpOnly: boolean;
    sameSite: string; // default, lax, strict, none
    raw: string;
    unparsed: string[];
}

declare type CookieConfig = {
    name: string;
    value: string;
    path?: string;
    domain?: string;
    expires?: number;
    maxAge?: number;
    secure?: boolean;
    httpOnly?: boolean;
    sameSite?: string; // default, lax, strict, none
}

declare namespace env {

    const variables: { [key: string]: any; };
    const arguments: { [key: string]: string; };
    const log: Logger;

    function loadString(id: string, timeoutSeconds?: number): string;
    function loadBytes(id: string, timeoutSeconds?: number): any;
    function writeFrom(writer: any, id: string, timeoutSeconds?: number): void;

}

declare namespace util {

    function stringToBytes(s: string): Bytes;
    function bytesToString(bytes: Bytes): string;
    function btoa(bytes: Bytes): string;
    function atob(b64: string): Bytes;
    function deepCopy(value: any): any;
    function deepEquals(a: any, b: any): boolean;
    function isType(value: any, typeName: string): boolean;
    function url(config: {
        scheme?: string;
        username?: string;
        password?: string;
        host?: string;
        port?: number;
        path?: string;
        query?: { [key: string]: string | string[]; };
        fragment?: string;
    }): string;
    function escapeHtml(s: string): string;
    function unescapeHtml(s: string): string;
    function hash(value: any): number;
    function sprintf(f: string, ...args: any[]): string;
    function fail(m: string): void;
    function failf(f: string, ...args: any[]): void;
    function now(): Time;
    function nowString(): string;
    function mutex(): Mutex;
    function once(name: string, f: () => void): void;
    function go(f: () => void): void;

    interface Time {} // TODO

    interface Mutex {
        lock(): void;
        unlock(): void;
    }

}

declare namespace transcribe {

    function write(writer: Writer, value: any, format: Format, indent?: string): void;
    function print(value: any, format: Format, indent?: string): void;
    function eprint(value: any, format: Format, indent?: string): void;
    function stringify(value: any, format: Format, indent?: string): string;
    function newXmlDocument(): XMLDocument;

    type Format = 'yaml' | 'json' | 'xjson' | 'xml' | 'cbor' | 'messagepack' | 'go'

    interface XMLDocument {} // TODO

}

declare namespace ard {

    function decode(bytes: Bytes, format: Format, all?: boolean): any;
    function validateFormat(bytes: Bytes, format: Format): void;

    type Format = 'yaml' | 'json' | 'xjson' | 'xml' | 'cbor' | 'messagepack'

}

declare namespace os {

    function joinFilePath(...elements: any[]): string;
    function exec(name: string, ...args: any[]): string;
    function temporaryFile(pattern: string, directory?: string): string;
    function temporaryDirectory(pattern: string, directory?: string): string;
    function download(sourceUrl: string, targetPath: string, timeoutSeconds?: number): void;

}

declare namespace scriptlet {

    function render(writer: Writer, content: any, renderer: string): void;
    function renderFrom(writer: Writer, id: string, renderer: string): void;
    function renderToString(content: any, renderer: string): string;

}

declare namespace prudence {

    const dataContentTypes: string[];
    const notFound: HandleFunction;
    const redirectTrailingSlash: HandleFunction;

    function start(startables: Startable | Startable[]): void;
    function setCache(backend: CacheBackend): void;
    function invalidateCacheGroup(group: string): void;
    function setScheduler(scheduler: Scheduler): void;
    function schedule(cronPattern: string, f: () => void): void;

    interface CacheBackend {}
    
    class TieredCache implements CacheBackend {
        constructor(config?: {
            caches?: CacheBackend[];
        });
    }

    class MapCache implements CacheBackend {
        constructor(config?: {
            pruneFrequency?: number;
        });
    }

    class MemoryCache implements CacheBackend {
        constructor(config?: {
            maxSize?: number;
            averageSize?: number;
            pruneFrequency?: number;
        });
    }

    class DistributedCache implements CacheBackend {
        constructor(config: {
            local: CacheBackend;
            kubernetes?: {
                namespace?: string;
                selector?: string;
            };
        });
    }

    interface Scheduler {} // TODO

    class LocalScheduler implements Scheduler {
        constructor(config?: {});
    }

    interface Startable {
        start(): void;
        stop(): void;
    }

    class Server implements Startable {
        constructor(config?: {
            name?: string;
            address?: string;
            port?: number;
            protocol?: 'dual' | 'ipv6' | 'ipv4';
            tls?: {
                certificate?: string;
                key?: string;
                generate?: boolean;
            };
            ncsaLogFileSuffix?: string;
            debug?: boolean;
            handlerTimeout?: number;
            readHeaderTimeout?: number;
            readTimeout?: number;
            writeTimeout?: number;
            idleTimeout?: number;
            handler?: Handler | HandleFunction;
        });

        start(): void;
        stop(): void;
    }

    type HandleFunction = () => boolean;

    interface Handler {
        handle: HandleFunction;
    }

    class Router implements Handler {
        constructor(config?: {
            name?: string;
            variables?: { [key: string]: any; };
            routes?: RouteConfig | RouteConfig[];
        });

        handle: HandleFunction;
    }

    class Route implements Handler {
        constructor(config?: RouteConfig);

        handle: HandleFunction;
    }

    type RouteConfig = {
        name?: string;
        paths?: string | string[];
        redirectTrailingSlashStatus?: number;
        variables?: { [key: string]: any; };
        handler?: Handler | HandleFunction;
    };

    class Resource implements Handler {
        constructor(config?: {
            name?: string;
            variables?: { [key: string]: any; };
            facets?: FacetConfig | FacetConfig[];
        });

        handle: HandleFunction;
    }

    class Facet implements Handler {
        constructor(config?: FacetConfig);

        handle: HandleFunction;
    }

    type FacetConfig = {
        name?: string;
        paths?: string | string[];
        redirectTrailingSlashStatus?: number;
        variables?: { [key: string]: any; };
        representations?: RepresentationConfig | RepresentationConfig[];
    };

    class Representation implements Handler {
        constructor(config?: RepresentationConfig);

        handle: HandleFunction;
    }

    type RepresentationHook = () => void;

    type RepresentationConfig = {
        name?: string;
        contentTypes?: string | string[];
        languages?: string | string[];
        charSet?: string;
        redirectTrailingSlash?: boolean;
        redirectTrailingSlashStatus?: number;
        variables?: { [key: string]: any; };
        prepare?: RepresentationHook;
        describe?: RepresentationHook;
        present?: RepresentationHook;
        erase?: RepresentationHook;
        modify?: RepresentationHook;
        call?: RepresentationHook;
        hooks?: {
            prepare?: RepresentationHook;
            describe?: RepresentationHook;
            present?: RepresentationHook;
            erase?: RepresentationHook;
            modify?: RepresentationHook;
            call?: RepresentationHook;
        };
    };

    class Static implements Handler {
        constructor(config?: {
            root?: string;
            indexes?: string[];
            presentDirectories?: boolean;
        });

        handle: HandleFunction;
    }

}
