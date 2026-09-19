const express = require("express");
const { connectDb } = require("./database/db");
const { userRoutes } = require("./modules/users/users.routes");
const { hashPassword } = require("./common/crypto");

const app = express();
app.use(express.json());

app.get("/health", (_req, res) => res.json({ ok: true }));
app.use("/api/users", userRoutes);

const PORT = process.env.PORT || 4000;
connectDb()
  .then(() => app.listen(PORT, () => console.log(`api on :${PORT}`)))
  .catch((err) => {
    console.error("db connection failed", err);
    process.exit(1);
  });
