// Shared DTOs — previously duplicated between client and server; both
// apps now import from this package so the shapes can never drift.

export type UserDto = {
  id: string;
  email: string;
  name: string;
};

export type SignupRequest = {
  email: string;
  name: string;
  password: string;
};
