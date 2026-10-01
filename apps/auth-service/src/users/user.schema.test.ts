import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { model } from "mongoose";
import { User, UserSchema } from "./user.schema";

describe("User JSON serialization", () => {
    it("keeps the public id while hiding password hash and timestamps", () => {
        const TestUser = model<User>("AuthServiceUserSerializationTest", UserSchema);
        const user = new TestUser({
            _id: "0123456789abcdef01234567",
            email: "reader@example.com",
            passwordHash: "secret-hash",
            createdAt: new Date("2026-01-01T00:00:00.000Z"),
            updatedAt: new Date("2026-01-02T00:00:00.000Z"),
        });

        assert.deepEqual(user.toJSON(), { email: "reader@example.com", id: "0123456789abcdef01234567" });
    });
});
