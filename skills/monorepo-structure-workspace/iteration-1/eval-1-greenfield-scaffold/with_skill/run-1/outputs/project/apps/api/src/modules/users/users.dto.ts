// The users contract lives in @repo/types so apps/web imports the same
// shapes the API serves. Re-exported here for NestJS's *.dto.ts convention.
export type { CreateUserDto, UpdateUserDto, UserDto } from "@repo/types";
