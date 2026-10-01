import { Prop, Schema, SchemaFactory } from "@nestjs/mongoose";
import { HydratedDocument } from "mongoose";

export type UserDocument = HydratedDocument<User>;

@Schema({
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
})
export class User {
    @Prop({ type: String, required: true, unique: true, lowercase: true, trim: true })
    email!: string;

    @Prop({ type: String, required: true })
    passwordHash!: string;
}

export const UserSchema = SchemaFactory.createForClass(User);
