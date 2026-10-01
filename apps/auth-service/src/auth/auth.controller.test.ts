import "reflect-metadata";
import assert from "node:assert/strict";
import { after, before, describe, it } from "node:test";
import { UnauthorizedException, type INestApplication } from "@nestjs/common";
import { Test } from "@nestjs/testing";
import { AuthController } from "./auth.controller";
import { AuthService } from "./auth.service";

describe("AuthController API", () => {
    let app: INestApplication;
    let baseUrl: string;

    before(async () => {
        const module = await Test.createTestingModule({
            controllers: [AuthController],
            providers: [
                {
                    provide: AuthService,
                    useValue: {
                        register: async (email: string) => ({ token: "registration-token", user: { email } }),
                        login: async (email: string, password: string) => {
                            if (email !== "reader@example.com" || password !== "password") {
                                throw new UnauthorizedException({ message: "Invalid credentials" });
                            }
                            return { token: "login-token", user: { email } };
                        },
                    },
                },
            ],
        }).compile();
        app = module.createNestApplication();
        app.setGlobalPrefix("api");
        await app.listen(0);
        const address = app.getHttpServer().address();
        assert.ok(address && typeof address === "object");
        baseUrl = `http://127.0.0.1:${address.port}`;
    });

    after(async () => {
        await app.close();
    });

    it("preserves the register endpoint response and status", async () => {
        const response = await fetch(`${baseUrl}/api/auth/register`, {
            method: "POST",
            headers: { "content-type": "application/json" },
            body: JSON.stringify({ email: "reader@example.com", password: "password" }),
        });

        assert.equal(response.status, 201);
        assert.deepEqual(await response.json(), {
            token: "registration-token",
            user: { email: "reader@example.com" },
        });
    });

    it("preserves required credential validation", async () => {
        const response = await fetch(`${baseUrl}/api/auth/register`, {
            method: "POST",
            headers: { "content-type": "application/json" },
            body: JSON.stringify({ email: "reader@example.com" }),
        });

        assert.equal(response.status, 400);
        assert.deepEqual(await response.json(), { message: "Email and password required" });
    });

    it("returns login success with HTTP 200", async () => {
        const response = await fetch(`${baseUrl}/api/auth/login`, {
            method: "POST",
            headers: { "content-type": "application/json" },
            body: JSON.stringify({ email: "reader@example.com", password: "password" }),
        });

        assert.equal(response.status, 200);
        assert.deepEqual(await response.json(), {
            token: "login-token",
            user: { email: "reader@example.com" },
        });
    });

    it("returns invalid credentials for missing login fields", async () => {
        const response = await fetch(`${baseUrl}/api/auth/login`, {
            method: "POST",
            headers: { "content-type": "application/json" },
            body: JSON.stringify({}),
        });

        assert.equal(response.status, 401);
        assert.deepEqual(await response.json(), { message: "Invalid credentials" });
    });
});
