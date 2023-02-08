import { api } from "../../lib/agent";
import type { APIRoute } from "../../lib/api/response";

export const get: APIRoute = async () => {
    return api("get", "/health", undefined, { Authorization: "Bearer" });
};

export const post: APIRoute = async ({ request }) => {
    // get url params from request
    return api(request.method, "/health", request.body, { debug: "fake authentication" });
};

