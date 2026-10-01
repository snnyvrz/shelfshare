import { CanActivate, ExecutionContext, Inject, Injectable, UnauthorizedException } from "@nestjs/common";
import { JwtService } from "@nestjs/jwt";

export type AuthenticatedUser = {
    id: string;
    email: string;
};

export type AuthenticatedRequest = {
    header(name: string): string | undefined;
    user?: AuthenticatedUser;
};

@Injectable()
export class JwtAuthGuard implements CanActivate {
    constructor(@Inject(JwtService) private readonly jwt: JwtService) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
        const request = context.switchToHttp().getRequest<AuthenticatedRequest>();
        const authHeader = request.header("authorization");
        if (!authHeader) {
            throw new UnauthorizedException({ error: "Missing Authorization header" });
        }

        const [scheme, token] = authHeader.split(" ");
        if (scheme !== "Bearer" || !token) {
            throw new UnauthorizedException({ error: "Invalid Authorization format" });
        }

        let payload: { sub?: unknown; email?: unknown };
        try {
            payload = await this.jwt.verifyAsync<{ sub?: unknown; email?: unknown }>(token);
        } catch {
            throw new UnauthorizedException({ error: "Invalid or expired token" });
        }

        if (
            typeof payload !== "object" ||
            payload === null ||
            !("sub" in payload) ||
            typeof payload.sub !== "string" ||
            !("email" in payload) ||
            typeof payload.email !== "string"
        ) {
            throw new UnauthorizedException({ error: "Invalid token payload" });
        }

        request.user = { id: payload.sub, email: payload.email };
        return true;
    }
}
