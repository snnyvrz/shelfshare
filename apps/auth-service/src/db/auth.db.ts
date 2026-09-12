import mongoose from "mongoose";

const createMongoUri = () => {
    const username = process.env.MONGO_INITDB_ROOT_USERNAME;
    const password = process.env.MONGO_INITDB_ROOT_PASSWORD;
    const host = process.env.MONGO_HOST || "localhost";
    const port = process.env.MONGO_PORT || "27017";
    const dbName = process.env.MONGO_INITDB_DATABASE || "shelfshare";

    if (username && password) {
        return `mongodb://${username}:${password}@${host}:${port}/${dbName}?authSource=admin`;
    } else {
        return `mongodb://${host}:${port}/${dbName}`;
    }
};

const MONGODB_URI = createMongoUri();

if (!MONGODB_URI) {
    throw new Error("MONGODB_URI is not defined");
}

let conn: typeof mongoose | null = null;

export async function connectMongo() {
    if (conn) return conn;

    try {
        conn = await mongoose.connect(MONGODB_URI, {
            autoIndex: true,
        });
        console.log("Connected to MongoDB");
        return conn;
    } catch (err) {
        console.error("MongoDB connection error:", err);
        conn = null;
        throw err;
    }
}
