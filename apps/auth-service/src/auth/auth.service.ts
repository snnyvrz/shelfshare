import {
    BadRequestException,
    ConflictException,
    Inject,
    Injectable,
    NotFoundException,
    UnauthorizedException,
} from "@nestjs/common";
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

    async register(email: string, password: string, displayName = "Reader") {
        const existing = await this.users.findOne({ email });
        if (existing) {
            throw new ConflictException({ message: "User already exists" });
        }

        const passwordHash: string = await argon2.hash(password, BUN_ARGON2_OPTIONS);
        let user: UserDocument;
        try {
            user = await this.users.create({ email, passwordHash, displayName });
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

    async profile(id: string) {
        if (!/^[a-f\d]{24}$/i.test(id)) throw new NotFoundException("Profile not found");
        const user = await this.users.findById(id);
        if (!user) throw new NotFoundException("Profile not found");
        return {
            id: user._id.toString(),
            displayName: user.displayName || "Reader",
            bio: user.bio || "",
            location: user.location || "",
        };
    }

    async profiles(query: string, page: number) {
        const escaped = query
            .trim()
            .slice(0, 100)
            .replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        const filter = escaped
            ? {
                  $or: [
                      { displayName: { $regex: escaped, $options: "i" } },
                      { location: { $regex: escaped, $options: "i" } },
                  ],
              }
            : {};
        const [users, total] = await Promise.all([
            this.users
                .find(filter)
                .select("_id displayName bio location")
                .sort({ displayName: 1, _id: 1 })
                .skip((page - 1) * 50)
                .limit(50)
                .lean(),
            this.users.countDocuments(filter),
        ]);
        return {
            data: users.map((user) => ({
                id: user._id.toString(),
                displayName: user.displayName || "Reader",
                bio: user.bio || "",
                location: user.location || "",
            })),
            total,
        };
    }

    async updateProfile(id: string, body: Record<string, unknown>) {
        const fields: Record<string, string> = {};
        for (const [key, limit] of [
            ["displayName", 80],
            ["bio", 1000],
            ["location", 120],
        ] as const) {
            if (body[key] === undefined) continue;
            if (
                typeof body[key] !== "string" ||
                body[key].length > limit ||
                (key === "displayName" && !body[key].trim())
            ) {
                throw new BadRequestException(`Invalid ${key}`);
            }
            fields[key] = body[key].trim();
        }
        await this.users.findByIdAndUpdate(id, { $set: fields }, { runValidators: true });
        return this.profile(id);
    }
}

function isDuplicateKeyError(error: unknown): error is { code: number } {
    return typeof error === "object" && error !== null && "code" in error && error.code === 11000;
}
