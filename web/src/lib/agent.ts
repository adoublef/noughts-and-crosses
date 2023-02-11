// https://web.dev/fetch-api-error-handling/

export const User = {
    create: (token: string, email: string, username: string, bio?: string) => send<{ location: string; username: string; }>("post", "/registry/users", { email, username, bio }, { Authorization: `Bearer ${token}` }),
    signup: {
        attempt: (email: string) => send<{ provider: string; }>("post", "/registry/signup", { email }),
        confirm: (token: string | null = "") => send<{ email: string; }>("get", "/registry/signup", undefined, { Authorization: `Bearer ${token}` }),
    },
} as const;

export const Auth = {
    login: {
        attempt: (email: string) => send<{ provider: string; }>("post", "/auth/login", { email }),
        confirm: (token: string | null = "") => send<{ username: string; }>("get", "/auth/login", undefined, { Authorization: `Bearer ${token}` }),
    },

} as const;

export const Ping = {
    hello: (hello: string = "") => send<{ sum: number; }>("post", "/health", { hello }),
} as const;

/** 
 * Does not throw an error, instead returns a 500 response.
 * My API should always return a json object, then can clean this up.
 */

async function send<T>(method: Method, url: string, payload?: any, headers: Record<string, string> = {}): Promise<SendError | T> {
    let opts: RequestInit = { method, headers, mode: "cors" };

    if (payload) {
        headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(payload);
    }

    // /*
    const controller = new AbortController();
    opts.signal = controller.signal;

    // Cancel the fetch request in 1000ms
    // NOTE: add a timeout to function arguments
    setTimeout(() => controller.abort(), 1000);
    // */

    try {
        const response = await fetch(import.meta.env.PUBLIC_API_URI + url, opts);
        if (!response.ok) {
            throw new Error("fetch error", { cause: response });
        }
        // 
        return (await response.json()) as T;
    } catch (err) {
        // SyntaxError (?) JSON parsing error
        if (err instanceof SyntaxError) {
            // TODO handle syntax error
            return { error: 1, name: "JSON parsing error" };
        };

        // DOMException: The user aborted a request. (AbortController)
        if (err instanceof DOMException) {
            // TODO handle syntax error
            return { error: 2, name: "Request aborted" };
        };

        // SystemError: connect ECONNREFUSED
        if (isSystemError(err)) {
            // TODO handle syntax error
            return { error: 3, name: "Connection refused" };
        };

        const { cause } = (err as Error);
        if (!isResponse(cause)) return { error: -1, name: "Unknown error" };

        switch (cause.status) {
            // TODO handle codes
            default:
                return { error: cause.status, name: cause.statusText };
        }
    }
}

type SendError = {
    error: number;
    name: string;
};

export function isSendError(err: any): err is SendError {
    return err?.error !== undefined;
}

export function isResponse(err: any): err is Response {
    return err?.status !== undefined;
}

/** 
 * This is a rough representation of the NodeJS's SystemError 
 * 
 * This does not include all the properties of the SystemError, 
 * but only the ones that can identify the error.
 */
type SystemError = {
    errno: -111,
    code: 'ECONNREFUSED';
    syscall: 'connect',
};

export function isSystemError(err: any): err is SystemError {
    return err?.code === "ECONNREFUSED";
}

const methods = ["get", "post", "put", "delete"] as const;
export type Method = typeof methods[number];

