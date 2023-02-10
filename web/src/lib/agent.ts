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
 * Does not throw an error, instead returns a 500 response
 * My API should always return a json object, then can clean this up
 */
async function send<T>(method: Method, url: string, payload?: unknown, headers: Record<string, string> = {}) {
    let opts: RequestInit = { method, headers, mode: "cors" };

    if (payload) {
        headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(payload);
    }

    try {
        const response = await fetch(import.meta.env.PUBLIC_API_URI + url, opts);
        if (!response.ok) {
            throw new Error("fetch error", { cause: response });
        }
        // 
        return (await response.json()) as T;
    } catch (err) {
        const { cause } = (err as Error);
        if (!isResponse(cause)) return { error: -1 } as const;
        switch (cause.status) {
            default:
                return { error: cause.status } as const;
        }
    }
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

