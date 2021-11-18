
declare function bind(id: string, name?: string): any;

declare interface Logger {
    critical(m: string): void
    criticalf(f: string, ...args: any[]): void
    error(m: string): void
    errorf(f: string, ...args: any[]): void
    warning(m: string): void
    warningf(f: string, ...args: any[]): void
    notice(m: string): void
    noticef(f: string, ...args: any[]): void
    info(m: string): void
    infof(f: string, ...args: any[]): void
    debug(m: string): void
    debugf(f: string, ...args: any[]): void
}

declare interface Context {
    request: ContextRequest;
    response: ContextResponse;
    log: Logger;
    name: string;
    debug: boolean;
    path: string;
    variables: { [key: string]: any; };
    done: boolean;
    created: boolean;
    async: boolean;
    cacheDuration: number;
    cacheKey: string;
    cacheGroups: string[];

    write(bytes: any): void;
    writeString(value: string): void;
    writeJson(value: any, indent?: string): void;
    writeYaml(value: any, indent?: string): void;
    redirect(url: string, status: number): void;
}

declare interface ContextRequest {
    host: string;
    port: number;
    header: Header;
    method: string;
    query: { [key: string]: string[]; };
    cookies: Cookie[];
    body: string;
    direct: any;
}

declare interface ContextResponse {
    status: number;
    header: Header;
    cookies: Cookie[];
    contentType: string;
    charSet: string;
    language: string;
    signature: string;
    weakSignature: boolean;
    timestamp: number;
    buffer: any;
    bypass: boolean;
    direct: any;

    reset(): void;
    addCookie(cookie: CookieConfig): void;
}

declare interface Header {
    add(key: string, value: string): void;
    set(key: string, value: string): void;
    get(key: string): string;
    values(key: string): string[];
}

type Cookie = {
    name: string;
    value: string;
    path: string;
    domain: string;
    expires: any;
    maxAge: number;
    secure: boolean;
    httpOnly: boolean;
    sameSite: number;
    raw: string;
    unparsed: string[];
}

type CookieConfig = {
    name: string;
    value: string;
    path?: string;
    domain?: string;
    expires?: number;
    maxAge?: number;
    secure?: boolean;
    httpOnly?: boolean;
    sameSite?: string;
}

declare namespace prudence {

    const arguments: { [key: string]: string; };
    const globals: { [key: string]: any; };
    const log: Logger;
    const defaultNotFound: Handler;

    // Platform
    function start(startables: Startable|Startable[]): void;
    function setCache(backend: CacheBackend): void;
    function invalidateCacheGroup(group: string): void;
    function setScheduler(scheduler: Scheduler): void;
    function schedule(cronPattern: string, f: () => void): void;
    function render(text: string, renderer: string): string;

    // Util
    function escapeHtml(text: string): string;
    function unescapeHtml(text: string): string;
    function stringToBytes(s: string): any;
    function bytesToString(bytes: any): string;
    function btoa(bytes: any): string;
    function atob(base64: string): any;
    function deepCopy(value: any): any;
    function deepEquals(a: any, b: any): boolean;
    function isType(value: any, type: string): boolean;
    function hash(value: any): string;
    function sprintf(f: string, ...args: any[]): string;
    function now(): any;
    function mutex(): Mutex;
    function once(name: string, f: () => void): void;
    function go(f: () => void): void;

    // Format
    function validateFormat(code: string, format: string): void;
    function decode(code: string, format: string, all?: boolean): any;
    function encode(value: any, format: string, indent?: string, writer?: any): void;
    function newXmlDocument(): any; // TODO

    // File
    function loadString(id: string): string;
    function loadBytes(id: string): any;
    function joinFilePath(...elements: any[]): string;
    function exec(name: string, ...args: any[]): string;
    function temporaryFile(pattern: string, directory?: string): string;
    function temporaryDirectory(pattern: string, directory?: string): string;
    function download(sourceUrl: string, targetPath: string): void;

    interface Mutex {
        lock(): void;
        unlock(): void;
    }

    interface CacheBackend {
    }

    class MemoryCache implements CacheBackend {
        constructor(config?: {});
    }

    class DistributedCache implements CacheBackend {
        constructor(config?: {});
    }

    interface Scheduler {
    }

    class QuartzScheduler implements Scheduler {
        constructor(config?: {});
    }

    interface Startable {
        start(): void;
        stop(): void;
    }

    class Server implements Startable {
        constructor(config?: {
            address?: string;
            handler?: Handler|HandleFunction;
        });

        start(): void;
        stop(): void;
    }

    type HandleFunction = (context: Context) => boolean;

    interface Handler {
        handle: HandleFunction;
    }

    class Router implements Handler {
        constructor(config?: {
            name?: string;
            routes?: RouteConfig|RouteConfig[];
        });

        handle: HandleFunction;
    }

    type RouteConfig = {
        name?: string;
        paths?: string|string[];
        handler?: Handler|HandleFunction;
    };

    class Resource implements Handler {
        constructor(config?: {
            name?: string;
            facets?: FacetConfig|FacetConfig[];
        });

        handle: HandleFunction;
    }

    type FacetConfig = {
        name?: string;
        paths?: string|string[];
        representations?: RepresentationConfig|RepresentationConfig[];
    };

    type RepresentationConfig = {
        functions?: {
            construct?: HandleFunction;
            describe?: HandleFunction;
            present?: HandleFunction;
            erase?: HandleFunction;
            modify?: HandleFunction;
            call?: HandleFunction;
        }|object;
        construct?: HandleFunction;
        describe?: HandleFunction;
        present?: HandleFunction;
        erase?: HandleFunction;
        modify?: HandleFunction;
        call?: HandleFunction;
    };

    class Static implements Handler {
        constructor(config?: {
            root?: string;
            indexes?: string[];
        });

        handle: HandleFunction;
    }

}
