export const User = {
    create: (token: string, email: string, username: string, bio?: string) => send<{ location: string; username: string; }>("post", "/registry/v0/users", { email, username, bio }, { Authorization: `Bearer ${token}` }),
};

export const Auth = {
    login: (email: string) => send<{ provider: string; }>("post", "/auth/v0/login", { email }),
    loginVerify: (token: string | null = "") => send<{username:string}>("get", "/auth/v0/login", undefined, { Authorization: `Bearer ${token}` }),
    signup: (email: string) => send<{ provider: string; }>("post", "/registry/v0/signup", { email }),
    signupVerify: (token: string | null = "") => send<{ email: string; }>("get", "/registry/v0/signup", undefined, { Authorization: `Bearer ${token}` }),
};

/** 
 * Does not throw an error, instead returns a 500 response
 * My API should always return a json object, then can clean this up
 */
async function send<T>(method: "post" | "get" | "delete", url: string, payload?: unknown, headers: Record<string, string> = {}) {
    let opts: RequestInit = { method, headers, mode: "cors" };

    if (payload) {
        headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(payload);
    }

    try {
        const response = await fetch(import.meta.env.PUBLIC_API_URI + url, opts);
        console.log(response.status, response.ok);
        if (!response.ok) {
            throw new Error("Bad fetch response", {
                cause: response,
            });
        }
        return (await response.json()) as T;
    } catch (err) {
        // const response = (err as Error).cause as Response;
        // switch (response.status) {
        //     default:
        //         return null;
        //         // return { err: true };
        //     }
        return null;
    }
}

async function parse<T extends Record<string, any>>(promise: Promise<Response>) {
    const res = await promise;
    switch (res.status) {
        case 204:
            return null;
        default:
            return res.json() as Promise<T>;
    }
}