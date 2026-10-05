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
            delete ret.discoveryCityId;
            delete ret.discoveryEnabled;
        },
    },
})
export class User {
    @Prop({ type: String, default: "Reader", trim: true, maxlength: 80 })
    displayName!: string;

    @Prop({ type: String, default: "", maxlength: 1000 })
    bio!: string;

    @Prop({ type: String, default: "", maxlength: 120 })
    location!: string;

    @Prop({ type: String, default: "" })
    discoveryCityId!: string;

    @Prop({ type: Boolean, default: false })
    discoveryEnabled!: boolean;

    @Prop({ type: String, required: true, unique: true, lowercase: true, trim: true })
    email!: string;

    @Prop({ type: String, required: true })
    passwordHash!: string;
}

export const UserSchema = SchemaFactory.createForClass(User);
UserSchema.index({ discoveryEnabled: 1, discoveryCityId: 1 });
