import { PrismaClient } from '@prisma/client';

export { PrismaClient } from '@prisma/client';
export { Prisma } from '@prisma/client';
export * from '@prisma/client';

/**
 * Shared PrismaClient instance for the dashboard database.
 * The NestJS API composes this into a provider; scripts/tests can import it
 * directly. A single client per process is recommended by Prisma.
 */
export const prisma = new PrismaClient();
