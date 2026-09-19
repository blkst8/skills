// The users contract. apps/api serves these shapes and apps/web imports
// them, so the two can never drift apart — no copy-pasted types.

export type UserRole = "admin" | "member" | "viewer";

export interface UserDto {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  createdAt: string; // ISO 8601
}

export interface CreateUserDto {
  email: string;
  name: string;
  role?: UserRole;
}

export interface UpdateUserDto {
  name?: string;
  role?: UserRole;
}
