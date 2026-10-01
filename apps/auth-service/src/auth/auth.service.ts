import { ConflictException, Inject, Injectable, UnauthorizedException } from "@nestjs/common";
import { InjectModel } from "@nestjs/mongoose";
import { JwtService } from "@nestjs/jwt";
import { Model } from "mongoose";
import * as argon2 from "argon2";
import { User, UserDocument } from "../users/user.schema";

const BUN_ARGON2_OPTIONS: argon2.HashOptions = {
    type: argon2.argon2id,
    memoryCost: 65536,
    timeCost: 2,
    parallelism: 1,
};

@Injectable()
export class AuthService {
    constructor(
        @InjectModel(User.name) private readonly users: Model<User>,
        @Inject(JwtService) private readonly jwt: JwtService
    ) {}

    async register(email: string, password: string) {
        const existing = await this.users.findOne({ email });
        if (existing) {
            throw new ConflictException({ message: "User already exists" });
        }

        const passwordHash: string = await argon2.hash(password, BUN_ARGON2_OPTIONS);
        let user: UserDocument;
        try {
            user = await this.users.create({ email, passwordHash });
        } catch (error) {
            if (isDuplicateKeyError(error)) {
                throw new ConflictException({ message: "User already exists" });
            }
            throw error;
        }

        return { token: await this.createToken(user), user };
    }

    async login(email: string, password: string) {
        const user = await this.users.findOne({ email });
        if (!user || !(await argon2.verify(user.passwordHash, password))) {
            throw new UnauthorizedException({ message: "Invalid credentials" });
        }

        return { token: await this.createToken(user), user };
    }

    private createToken(user: UserDocument) {
        return this.jwt.signAsync({ sub: user._id.toString(), email: user.email });
    }
}

function isDuplicateKeyError(error: unknown): error is { code: number } {
    return typeof error === "object" && error !== null && "code" in error && error.code === 11000;
}
