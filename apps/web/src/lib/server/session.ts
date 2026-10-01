import { jwtVerify } from "jose";

export const sessionCookie = "shelfshare_session";

export function safeReturnTo(value: string | null): string {
    return value && value.startsWith("/") && !value.startsWith("//") && !value.includes("\\") ? value : "/books";
}

export async function verifySession(token: string, secret: string) {
    if (!secret) return null;
    try {
        const { payload } = await jwtVerify(token, new TextEncoder().encode(secret), {
            algorithms: ["HS256"],
            requiredClaims: ["exp", "sub", "email"],
        });
        if (typeof payload.sub !== "string" || typeof payload.email !== "string") return null;
        return { user: { id: payload.sub, email: payload.email }, expires: new Date(payload.exp! * 1000) };
    } catch {
        return null;
    }
}
