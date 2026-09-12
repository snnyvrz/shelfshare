import { type HydratedDocumentFromSchema, Schema, model, type InferSchemaType } from "mongoose";

const userSchema = new Schema(
    {
        email: {
            type: String,
            required: true,
            unique: true,
            lowercase: true,
            trim: true,
        },
        passwordHash: {
            type: String,
            required: true,
        },
    },
    {
        timestamps: true,
        toJSON: {
            virtuals: true,
            versionKey: false,
            transform(_doc, ret: Record<string, unknown>) {
                delete ret._id;
                delete ret.passwordHash;
                delete ret.createdAt;
                delete ret.updatedAt;
            },
        },
    }
);

export type User = InferSchemaType<typeof userSchema>;

export type UserDoc = HydratedDocumentFromSchema<typeof userSchema>;

export const UserModel = model<UserDoc>("User", userSchema);
