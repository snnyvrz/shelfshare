import type { ConfigService } from "@nestjs/config";

export function createMongoUri(config: Pick<ConfigService, "get">): string {
    const username = config.get<string>("MONGO_INITDB_ROOT_USERNAME");
    const password = config.get<string>("MONGO_INITDB_ROOT_PASSWORD");
    const host = config.get<string>("MONGO_HOST") || "localhost";
    const port = config.get<string>("MONGO_PORT") || "27017";
    const dbName = config.get<string>("MONGO_INITDB_DATABASE") || "shelfshare";

    if (username && password) {
        const credentials = `${encodeURIComponent(username)}:${encodeURIComponent(password)}`;
        return `mongodb://${credentials}@${host}:${port}/${encodeURIComponent(dbName)}?authSource=admin`;
    }

    return `mongodb://${host}:${port}/${encodeURIComponent(dbName)}`;
}
