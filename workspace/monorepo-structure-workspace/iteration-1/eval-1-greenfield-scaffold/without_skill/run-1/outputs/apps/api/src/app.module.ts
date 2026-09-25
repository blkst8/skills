import { Module } from '@nestjs/common';
import { PrismaClient } from '@dashboard/database';
import { AppController } from './app.controller';
import { AppService } from './app.service';

@Module({
  imports: [],
  controllers: [AppController],
  providers: [
    AppService,
    // Single PrismaClient instance shared across the whole API.
    // Swap for a dedicated PrismaModule when the first feature module lands.
    { provide: PrismaClient, useValue: new PrismaClient() },
  ],
})
export class AppModule {}
