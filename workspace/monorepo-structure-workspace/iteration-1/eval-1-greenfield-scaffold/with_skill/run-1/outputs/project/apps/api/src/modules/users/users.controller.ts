import { Body, Controller, Get, Post } from "@nestjs/common";
import { API_ROUTES, createUserSchema } from "@repo/shared";
import type { CreateUserDto, UserDto } from "@repo/types";
import { UsersService } from "./users.service";

@Controller(API_ROUTES.users)
export class UsersController {
  constructor(private readonly users: UsersService) {}

  @Get()
  list(): UserDto[] {
    return this.users.list();
  }

  @Post()
  create(@Body() dto: CreateUserDto): UserDto {
    // Validate with the same schema the frontend validates its form with.
    const payload = createUserSchema.parse(dto);
    return this.users.create(payload);
  }
}
