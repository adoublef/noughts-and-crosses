// Helper for REST API calls
export function send() { }

export async function confirmEmail(token: string | null) {
    return fetch(
        `${import.meta.env.PUBLIC_API_URI}/auth/v0/login`,
        { headers: { Authorization: `Bearer ${token}` } }
    );
}

export async function login(email: string) {
    return fetch(
        `${import.meta.env.PUBLIC_API_URI}/auth/v0/login`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email }),
        }
    );
}