import assert from "node:assert/strict";
import { test } from "node:test";
import { SignJWT } from "jose";
import { safeReturnTo, verifySession } from "../src/lib/server/session";

test("session verification accepts auth-service claims and rejects invalid sessions", async () => {
    const secret = "shared-test-secret";
    const key = new TextEncoder().encode(secret);
    const sign = (algorithm: string, expiry?: string, email: unknown = "reader@example.com") => {
        let jwt = new SignJWT({ email }).setProtectedHeader({ alg: algorithm }).setSubject("user-id");
        if (expiry) jwt = jwt.setExpirationTime(expiry);
        return jwt.sign(key);
    };
    const valid = await sign("HS256", "1h");
    assert.deepEqual((await verifySession(valid, secret))?.user, { id: "user-id", email: "reader@example.com" });
    for (const token of [
        "broken",
        await sign("HS256", "-1h"),
        await sign("HS256"),
        await sign("HS384", "1h"),
        await sign("HS256", "1h", 42),
    ]) {
        assert.equal(await verifySession(token, secret), null);
    }
    assert.equal(await verifySession(valid, "wrong-secret"), null);
    assert.equal(await verifySession(valid, ""), null);
});

test("post-login destinations stay on the same site", () => {
    for (const value of [null, "https://example.com", "//example.com", "/\\example.com"])
        assert.equal(safeReturnTo(value), "/books");
    assert.equal(safeReturnTo("/books/new?from=catalog"), "/books/new?from=catalog");
});
