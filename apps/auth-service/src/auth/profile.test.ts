import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { BadRequestException, NotFoundException } from "@nestjs/common";
import { JwtService } from "@nestjs/jwt";
import type { Model } from "mongoose";
import { AuthService } from "./auth.service";
import type { User } from "../users/user.schema";

describe("Public profile privacy and updates", () => {
    const id = "0123456789abcdef01234567";
    it("uses an explicit public projection and supports existing accounts", async () => {
        const service = new AuthService(
            {
                findById: async () => ({ _id: id, email: "private@example.com", passwordHash: "secret" }),
            } as unknown as Model<User>,
            new JwtService()
        );
        assert.deepEqual(await service.profile(id), { id, displayName: "Reader", bio: "", location: "" });
        await assert.rejects(service.profile("invalid"), NotFoundException);
    });
    it("validates and whitelists profile changes", async () => {
        let update: unknown;
        const service = new AuthService(
            {
                findById: async () => ({ _id: id, displayName: "Reader" }),
                findByIdAndUpdate: async (_id: string, fields: unknown) => {
                    update = fields;
                },
            } as unknown as Model<User>,
            new JwtService()
        );
        await service.updateProfile(id, {
            displayName: " New name ",
            email: "changed@example.com",
            passwordHash: "bad",
        });
        assert.deepEqual(update, { $set: { displayName: "New name" } });
        for (const fields of [{ displayName: " " }, { bio: "x".repeat(1001) }, { location: 7 }])
            await assert.rejects(service.updateProfile(id, fields), BadRequestException);
    });
});
