import { API_ROUTES, createUserSchema } from "@repo/shared";
import type { CreateUserDto, UserDto } from "@repo/types";
import { apiFetch } from "@/lib/api-client";

export function listUsers(): Promise<UserDto[]> {
  return apiFetch<UserDto[]>(API_ROUTES.users);
}

export async function createUser(input: CreateUserDto): Promise<UserDto> {
  // Same schema the API enforces server-side — no drift possible.
  const payload = createUserSchema.parse(input);
  return apiFetch<UserDto>(API_ROUTES.users, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
