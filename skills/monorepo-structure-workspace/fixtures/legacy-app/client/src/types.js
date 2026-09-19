// NOTE: this type is duplicated on the server (server/routes.js returns
// these shapes inline) — the restructure should extract it to a shared
// package and have BOTH sides import from it.

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
