const { Pool } = require("pg");

let pool;

function connectDb() {
  pool = new Pool({ connectionString: process.env.DATABASE_URL });
  return pool.query("SELECT 1");
}

function getPool() {
  if (!pool) throw new Error("db not connected");
  return pool;
}

module.exports = { connectDb, getPool };
