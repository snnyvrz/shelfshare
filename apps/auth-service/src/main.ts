import "reflect-metadata";
import { NestFactory } from "@nestjs/core";
import { AppModule } from "./app.module";

async function bootstrap() {
    const app = await NestFactory.create(AppModule);
    app.setGlobalPrefix("api");

    const port = Number(process.env.PORT ?? 3030);
    await app.listen(port);
    console.log(`Auth service listening on port ${port}`);
}

bootstrap().catch((err: unknown) => {
    console.error("Failed to start auth-service:", err);
    process.exit(1);
});
