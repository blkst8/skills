// Types shared between apps/web and apps/api so the client and server
// cannot drift apart. This file was extracted from client/src/types.js
// (previously duplicated between client and server).
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
