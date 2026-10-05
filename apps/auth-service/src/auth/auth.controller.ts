import {
    BadRequestException,
    Body,
    Controller,
    Get,
    HttpCode,
    HttpStatus,
    Inject,
    Param,
    Patch,
    Post,
    Query,
    Req,
    UseGuards,
} from "@nestjs/common";
import { AuthService } from "./auth.service";
import { JwtAuthGuard, type AuthenticatedRequest } from "./jwt-auth.guard";
import { cityOptions } from "../discovery/cities";

type Credentials = {
    email?: string;
    password?: string;
    displayName?: string;
};

@Controller("auth")
export class AuthController {
    constructor(@Inject(AuthService) private readonly auth: AuthService) {}

    @Post("register")
    async register(@Body() body: Credentials) {
        const { email, password } = body ?? {};
        if (typeof email !== "string" || !email.trim() || typeof password !== "string" || !password) {
            throw new BadRequestException({ message: "Email and password required" });
        }

        if (
            body.displayName !== undefined &&
            (typeof body.displayName !== "string" || !body.displayName.trim() || body.displayName.length > 80)
        ) {
            throw new BadRequestException("Display name must be 1–80 characters");
        }
        return body.displayName === undefined
            ? this.auth.register(email.trim().toLowerCase(), password)
            : this.auth.register(email.trim().toLowerCase(), password, body.displayName.trim());
    }

    @Post("login")
    @HttpCode(HttpStatus.OK)
    async login(@Body() body: Credentials) {
        const { email, password } = body ?? {};
        return this.auth.login(
            typeof email === "string" ? email.trim().toLowerCase() : "",
            typeof password === "string" ? password : ""
        );
    }

    @Get("profiles/:id")
    profile(@Param("id") id: string) {
        return this.auth.profile(id);
    }

    @Get("profiles")
    profiles(@Query("q") query = "", @Query("page") page = "1") {
        const number = Number(page);
        return this.auth.profiles(
            typeof query === "string" ? query : "",
            Number.isInteger(number) && number > 0 ? Math.min(number, 100000) : 1
        );
    }

    @Get("me")
    @UseGuards(JwtAuthGuard)
    me(@Req() request: AuthenticatedRequest) {
        return this.auth.profile(request.user!.id, true);
    }

    @Get("cities")
    cities(@Query("q") query = "") {
        if (typeof query !== "string") throw new BadRequestException("Invalid city search");
        return { data: cityOptions(query) };
    }

    @Get("nearby/profiles")
    @UseGuards(JwtAuthGuard)
    nearbyProfiles(@Query() params: Record<string, unknown>) {
        return this.auth.nearbyProfiles(params);
    }

    @Get("nearby/owners")
    @UseGuards(JwtAuthGuard)
    nearbyOwners(@Query() params: Record<string, unknown>) {
        return this.auth.nearbyOwners(params);
    }

    @Patch("me")
    @UseGuards(JwtAuthGuard)
    update(@Req() request: AuthenticatedRequest, @Body() body: Record<string, unknown>) {
        if (!body || typeof body !== "object" || Array.isArray(body)) throw new BadRequestException("Invalid profile");
        return this.auth.updateProfile(request.user!.id, body);
    }
}
