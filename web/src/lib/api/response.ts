import type { APIContext as AstroAPIContext, APIRoute as AstroAPIRoute, EndpointOutput as AstroEndpointOutput } from "astro";


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

function isSystemError(err: any): err is SystemError {
    return err?.code === "ECONNREFUSED";
}

const methods = ["get", "post", "put", "delete"] as const;
export type Method = typeof methods[number];

export type APIRoute = (context: APIContext) => AstroEndpointOutput | Response | Promise<AstroEndpointOutput | Response>;

export type APIContext =
    & Omit<AstroAPIContext, "request">
    & { request: Omit<Request, "method"> & { method: Method; }; };

