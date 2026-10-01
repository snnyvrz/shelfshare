import { Module } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import { JwtModule } from "@nestjs/jwt";
import { AuthController } from "./auth.controller";
import { AuthService } from "./auth.service";
import { JwtAuthGuard } from "./jwt-auth.guard";
import { UsersModule } from "../users/users.module";

@Module({
    imports: [
        UsersModule,
        JwtModule.registerAsync({
            inject: [ConfigService],
            useFactory: (config: ConfigService) => {
                const secret = config.get<string>("JWT_SECRET");
                if (!secret) {
                    throw new Error("JWT_SECRET environment variable is required");
                }

                const expiresIn = Number(config.get<string>("JWT_EXPIRES_IN") || 3600);
                if (!Number.isFinite(expiresIn) || expiresIn <= 0) {
                    throw new Error("JWT_EXPIRES_IN must be a positive number of seconds");
                }

                return { secret, signOptions: { expiresIn } };
            },
        }),
    ],
    controllers: [AuthController],
    providers: [AuthService, JwtAuthGuard],
    exports: [AuthService, JwtAuthGuard],
})
export class AuthModule {}
