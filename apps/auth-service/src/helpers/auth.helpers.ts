import jwt from "jsonwebtoken";

import type { User } from "@auth/types/auth.types";

const JWT_SECRET = process.env.JWT_SECRET || "dev_secret";
const JWT_EXPIRES_IN = process.env.JWT_EXPIRES_IN || 3600;

export const signToken = (user: User) => {
    return jwt.sign({ sub: user.id, email: user.email }, JWT_SECRET, {
        expiresIn: Number(JWT_EXPIRES_IN),
    });
};
