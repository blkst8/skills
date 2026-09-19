import { Global, Module } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";

// Database code lives here (schema, models, migrations, connection) — never
// in packages/. The frontend knows the world through DTOs in @repo/types.
export const DATABASE_URL = "DATABASE_URL";

@Global()
@Module({
  providers: [
    {
      provide: DATABASE_URL,
      useFactory: (config: ConfigService) => config.getOrThrow<string>("database.url"),
      inject: [ConfigService],
    },
  ],
  exports: [DATABASE_URL],
})
export class DatabaseModule {}
