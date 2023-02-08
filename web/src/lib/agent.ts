import { isResponse, Method } from "./api/response";

export const User = {
    create: (token: string, email: string, username: string, bio?: string) => send<{ location: string; username: string; }>("post", "/registry/v0/users", { email, username, bio }, { Authorization: `Bearer ${token}` }),
};

export const Auth = {
    login: (email: string) => send<{ provider: string; }>("post", "/auth/v0/login", { email }),
    loginVerify: (token: string | null = "") => send<{ username: string; }>("get", "/auth/v0/login", undefined, { Authorization: `Bearer ${token}` }),
    signup: (email: string) => send<{ provider: string; }>("post", "/registry/v0/signup", { email }),
    signupVerify: (token: string | null = "") => send<{ email: string; }>("get", "/registry/v0/signup", undefined, { Authorization: `Bearer ${token}` }),
};

export const Ping = {
    hello: (hello: string = "") => send<{ sum: number; }>("post", "/health", { hello }),
};

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
        return (await response.json()) as T;
    } catch (err) {
        const { cause } = (err as Error);
        if (!isResponse(cause)) {
            return { error: -1 } as const;
        }
        switch (cause.status) {
            default:
                return { error: cause.status } as const;
        }
    }
}

