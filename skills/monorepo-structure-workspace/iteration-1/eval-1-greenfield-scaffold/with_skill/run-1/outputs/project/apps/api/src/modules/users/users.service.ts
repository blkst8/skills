import { Injectable } from "@nestjs/common";
import { randomUUID } from "node:crypto";
import type { CreateUserDto, UserDto } from "@repo/types";

@Injectable()
export class UsersService {
  // In-memory until the Postgres-backed repository lands with the first
  // real feature — the module boundary means callers never notice.
  private readonly users: UserDto[] = [];

  list(): UserDto[] {
    return this.users;
  }

  create(dto: CreateUserDto): UserDto {
    const user: UserDto = {
      id: randomUUID(),
      email: dto.email,
      name: dto.name,
      role: dto.role ?? "member",
      createdAt: new Date().toISOString(),
    };
    this.users.push(user);
    return user;
  }
}
