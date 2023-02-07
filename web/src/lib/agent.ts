export const Auth = {
    /** I should be able to return the response body of the request in a type-safe way */
    loginVerify: (token: string | null = "") => send("get", "/auth/v0/login", undefined, { Authorization: `Bearer ${token}` }),
    /** I should be able to return the response body of the request in a type-safe way */
    login: (email: string) => send("post", "/auth/v0/login", { email }),
    signup: (email: string, username: string) => send("post", "/registry/v0/signup", { email, username }),
    signupVerify: (token: string | null = "") => send("get", "/auth/v0/login", undefined, { Authorization: `Bearer ${token}` }),
};

/** 
 * Does not throw an error, instead returns a 500 response
 * My API should always return a json object, then can clean this up
 */
async function send(method: "post" | "get", url: string, payload?: any, headers: Record<string, string> = {}) {
    let opts: RequestInit = { method, headers };

    if (payload) {
        headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(payload);
    }

    try {
        return fetch(import.meta.env.PUBLIC_API_URI + url, opts);
    } catch (err) {
        return new Response(null, { status: 500 });
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