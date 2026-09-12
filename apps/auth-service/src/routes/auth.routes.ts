// src/routes/auth.ts
import { Router } from "express";
import { UserModel } from "@auth/models/User";
import { signToken } from "@auth/helpers/auth.helpers";

const router = Router();

router.post("/register", async (req, res) => {
    const { email, password } = req.body as { email: string; password: string };

    if (!email || !password) {
        return res.status(400).json({ message: "Email and password required" });
    }

    const existing = await UserModel.findOne({ email });
    if (existing) {
        return res.status(409).json({ message: "User already exists" });
    }

    const passwordHash = await Bun.password.hash(password);

    const user = await UserModel.create({
        email,
        passwordHash,
    });

    const token = signToken(user);

    return res.status(201).json({ token, user });
});

router.post("/login", async (req, res) => {
    const { email, password } = req.body as { email: string; password: string };

    const user = await UserModel.findOne({ email });
    if (!user) {
        return res.status(401).json({ message: "Invalid credentials" });
    }

    const match = await Bun.password.verify(password, user.passwordHash);
    if (!match) {
        return res.status(401).json({ message: "Invalid credentials" });
    }

    const token = signToken(user);

    return res.json({ token, user });
});

export { router };
