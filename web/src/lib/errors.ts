/** NOTE this can be turned into an AppError class */
export function isAppError(a: any): a is { error: number; } {
    return "error" in a;
}
