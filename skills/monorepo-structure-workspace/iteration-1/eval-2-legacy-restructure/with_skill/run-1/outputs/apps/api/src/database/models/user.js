const { getPool } = require("../db");

// NOTE: password hashing logic lives here with the model — must survive restructure
class UserModel {
  static async list() {
    const { rows } = await getPool().query("SELECT id, email, name FROM users");
    return rows;
  }

  static async findByEmail(email) {
    const { rows } = await getPool().query(
      "SELECT * FROM users WHERE email = $1",
      [email]
    );
    return rows[0] || null;
  }

  static async create({ email, name, passwordHash }) {
    const { rows } = await getPool().query(
      "INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id, email, name",
      [email, name, passwordHash]
    );
    return rows[0];
  }
}

module.exports = { UserModel };
