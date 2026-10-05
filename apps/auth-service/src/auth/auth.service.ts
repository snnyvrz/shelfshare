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
import { cityById, cityLabel, discoveryCities } from "../discovery/cities";

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

    async profile(id: string, owner = false) {
        if (!/^[a-f\d]{24}$/i.test(id)) throw new NotFoundException("Profile not found");
        const user = await this.users.findById(id);
        if (!user) throw new NotFoundException("Profile not found");
        return {
            id: user._id.toString(),
            displayName: user.displayName || "Reader",
            bio: user.bio || "",
            location: user.location || "",
            ...(owner
                ? { discoveryCityId: user.discoveryCityId || "", discoveryEnabled: user.discoveryEnabled === true }
                : {}),
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
        const fields: Record<string, string | boolean> = {};
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
        if (body.discoveryCityId !== undefined || body.discoveryEnabled !== undefined || body.location !== undefined) {
            const current = await this.users.findById(id);
            if (!current) throw new NotFoundException("Profile not found");
            const cityId = body.discoveryCityId === undefined ? current.discoveryCityId || "" : body.discoveryCityId;
            if (typeof cityId !== "string") throw new BadRequestException("Invalid discovery city");
            const city = cityId ? cityById(cityId) : null;
            const enabled =
                body.discoveryEnabled === undefined
                    ? body.discoveryCityId === ""
                        ? false
                        : current.discoveryEnabled === true
                    : body.discoveryEnabled;
            if (typeof enabled !== "boolean") throw new BadRequestException("Invalid discovery preference");
            if (enabled && !city) throw new BadRequestException("Select a city before enabling nearby discovery");
            if (body.discoveryCityId !== undefined) fields.discoveryCityId = cityId;
            if (body.discoveryEnabled !== undefined) fields.discoveryEnabled = enabled;
            else if (body.discoveryCityId === "") fields.discoveryEnabled = false;
            if (city) fields.location = cityLabel(city);
            else if (body.discoveryCityId === "" && current.discoveryCityId) fields.location = "";
        }
        await this.users.findByIdAndUpdate(id, { $set: fields }, { runValidators: true });
        return this.profile(id, true);
    }

    async nearbyProfiles(params: Record<string, unknown>) {
        const ranked = discoveryCities(params);
        const page = Number(params.page ?? "1");
        if (!Number.isInteger(page) || page < 1 || page > 100000) throw new BadRequestException("Invalid page");
        if (params.q !== undefined && typeof params.q !== "string") throw new BadRequestException("Invalid search");
        const q = String(params.q || "")
            .trim()
            .slice(0, 100)
            .replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        const result = await this.users.aggregate([
            {
                $match: {
                    discoveryEnabled: true,
                    discoveryCityId: { $in: ranked.map((city) => city.id) },
                    ...(q
                        ? {
                              $or: [
                                  { displayName: { $regex: q, $options: "i" } },
                                  { location: { $regex: q, $options: "i" } },
                              ],
                          }
                        : {}),
                },
            },
            { $addFields: { cityRank: { $indexOfArray: [ranked.map((city) => city.id), "$discoveryCityId"] } } },
            { $sort: { cityRank: 1, displayName: 1, _id: 1 } },
            {
                $facet: {
                    data: [
                        { $skip: (page - 1) * 50 },
                        { $limit: 50 },
                        { $project: { _id: 0, id: { $toString: "$_id" }, displayName: 1, bio: 1, location: 1 } },
                    ],
                    total: [{ $count: "count" }],
                },
            },
        ]);
        return { data: result[0]?.data || [], total: result[0]?.total[0]?.count || 0 };
    }

    // All eligible owners, independent of reader search/pagination. No coordinates leave auth-service.
    async nearbyOwners(params: Record<string, unknown>) {
        const ranked = discoveryCities(params);
        const users = await this.users
            .find({ discoveryEnabled: true, discoveryCityId: { $in: ranked.map((city) => city.id) } })
            .select("_id discoveryCityId")
            .lean();
        const ranks = new Map(ranked.map((city, index) => [city.id, index]));
        return {
            data: users.map((user) => ({ ownerId: user._id.toString(), rank: ranks.get(user.discoveryCityId)! })),
        };
    }
}

function isDuplicateKeyError(error: unknown): error is { code: number } {
    return typeof error === "object" && error !== null && "code" in error && error.code === 11000;
}
