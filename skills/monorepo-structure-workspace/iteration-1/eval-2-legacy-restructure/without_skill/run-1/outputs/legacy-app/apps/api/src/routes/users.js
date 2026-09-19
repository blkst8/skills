const express = require("express");
const { UserModel } = require("../models/user.model");
const { hashPassword, verifyPassword } = require("../lib/crypto");

// Request/response shapes are shared with apps/web via @repo/shared
// (UserDto, SignupRequest) so client and server cannot drift apart.
/** @typedef {import('@repo/shared').UserDto} UserDto */
/** @typedef {import('@repo/shared').SignupRequest} SignupRequest */

const userRoutes = express.Router();

userRoutes.get("/", async (_req, res) => {
  const users = await UserModel.list();
  res.json(users);
});

userRoutes.post("/", async (req, res) => {
  const { email, name, password } = req.body;
  const existing = await UserModel.findByEmail(email);
  if (existing) return res.status(409).json({ error: "email taken" });
  const passwordHash = hashPassword(password);
  const user = await UserModel.create({ email, name, passwordHash });
  res.status(201).json(user);
});

userRoutes.post("/login", async (req, res) => {
  const { email, password } = req.body;
  const user = await UserModel.findByEmail(email);
  if (!user || !verifyPassword(password, user.passwordHash)) {
    return res.status(401).json({ error: "invalid credentials" });
  }
  res.json({ id: user.id, email: user.email, name: user.name });
});

module.exports = { userRoutes };
