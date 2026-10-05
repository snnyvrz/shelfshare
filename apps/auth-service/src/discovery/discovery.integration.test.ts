import assert from "node:assert/strict";
import { it } from "node:test";
import { createConnection } from "mongoose";
import { JwtService } from "@nestjs/jwt";
import { BadRequestException } from "@nestjs/common";
import { AuthService } from "../auth/auth.service";
import { User, UserSchema } from "../users/user.schema";

// Only run against a disposable MongoDB. This test drops its dedicated database.
it(
    "discovers opted-in readers with real MongoDB ranking, totals and pagination",
    { skip: !process.env.DISCOVERY_MONGO_URL },
    async () => {
        const connection = await createConnection(process.env.DISCOVERY_MONGO_URL!, {
            dbName: "shelfshare_discovery_test",
        }).asPromise();
        try {
            const users = connection.model<User>("User", UserSchema);
            await users.init();
            await users.deleteMany({});
            const service = new AuthService(users, new JwtService());
            const make = (email: string, cityId: string, discoveryEnabled = true, displayName = "Reader") =>
                users.create({
                    email,
                    passwordHash: "private",
                    discoveryCityId: cityId,
                    discoveryEnabled,
                    displayName,
                    location: "City",
                    bio: "Books",
                });
            const karaj = await make("karaj@example.com", "128747", true, "A nearby reader");
            const tehran = await make("tehran@example.com", "112931", true, "Z origin reader");
            await make("off@example.com", "112931", false);
            await make("missing@example.com", "");
            await make("far@example.com", "124665");
            const params = { mode: "radius", cityId: "112931", radiusKm: "50" };
            const result = await service.nearbyProfiles(params);
            assert.equal(result.total, 2);
            assert.deepEqual(
                result.data.map((u: { id: string }) => u.id),
                [tehran.id, karaj.id]
            );
            assert.deepEqual(Object.keys(result.data[0]).sort(), ["bio", "displayName", "id", "location"]);
            assert.equal((await service.nearbyProfiles({ ...params, q: "nearby" })).total, 1);
            assert.equal((await service.nearbyProfiles({ ...params, mode: "city" })).total, 1);
            assert.equal((await service.nearbyOwners(params)).data.length, 2);
            await assert.rejects(service.nearbyProfiles({ ...params, page: "0" }), BadRequestException);
            for (const body of [
                { discoveryEnabled: "true" },
                { discoveryCityId: "bad" },
                { discoveryCityId: "", discoveryEnabled: true },
            ])
                await assert.rejects(service.updateProfile(tehran.id, body), BadRequestException);
            await service.updateProfile(tehran.id, { discoveryEnabled: false });
            assert.equal((await service.nearbyProfiles(params)).total, 1);
            const publicProfile = await service.profile(tehran.id);
            assert.equal("discoveryCityId" in publicProfile, false);
            const own = await service.updateProfile(tehran.id, {
                discoveryCityId: "128747",
                discoveryEnabled: true,
                location: "Override",
            });
            assert.equal(own.discoveryCityId, "128747");
            assert.ok(own.location.includes("Karaj"));
            await service.updateProfile(tehran.id, { discoveryCityId: "" });
            const cleared = await service.profile(tehran.id, true);
            assert.equal(cleared.discoveryEnabled, false);
            assert.equal(cleared.location, "");
            await Promise.all(
                Array.from({ length: 51 }, (_, i) =>
                    make(`page${i}@example.com`, "112931", true, `Reader ${String(i).padStart(2, "0")}`)
                )
            );
            const first = await service.nearbyProfiles({ mode: "city", cityId: "112931" });
            const second = await service.nearbyProfiles({ mode: "city", cityId: "112931", page: "2" });
            assert.equal(first.total, 51);
            assert.equal(first.data.length, 50);
            assert.equal(second.data.length, 1);
            assert.equal(new Set([...first.data, ...second.data].map((u: { id: string }) => u.id)).size, 51);
        } finally {
            await connection.dropDatabase();
            await connection.close();
        }
    }
);
