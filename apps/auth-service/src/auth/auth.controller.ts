import { BadRequestException, Body, Controller, HttpCode, HttpStatus, Inject, Post } from "@nestjs/common";
import { AuthService } from "./auth.service";

type Credentials = {
    email?: string;
    password?: string;
};

@Controller("auth")
export class AuthController {
    constructor(@Inject(AuthService) private readonly auth: AuthService) {}

    @Post("register")
    async register(@Body() body: Credentials) {
        const { email, password } = body ?? {};
        if (!email || !password) {
            throw new BadRequestException({ message: "Email and password required" });
        }

        return this.auth.register(email, password);
    }

    @Post("login")
    @HttpCode(HttpStatus.OK)
    async login(@Body() body: Credentials) {
        const { email, password } = body ?? {};
        return this.auth.login(email ?? "", password ?? "");
    }
}
