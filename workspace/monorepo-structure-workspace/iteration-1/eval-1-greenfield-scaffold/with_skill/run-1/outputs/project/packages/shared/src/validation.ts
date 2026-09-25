import { z } from "zod";

// The highest-value resident of this package: the API validates a request
// with the same schema the frontend validates its form with, so the two can
// never drift apart. These schemas validate the DTO shapes in @repo/types.

export const userRoleSchema = z.enum(["admin", "member", "viewer"]);

export const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
  role: userRoleSchema.optional(),
});

export const updateUserSchema = z.object({
  name: z.string().min(1).optional(),
  role: userRoleSchema.optional(),
});
