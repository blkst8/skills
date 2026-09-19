import React, { useState } from "react";
import { fetchUsers, signup } from "./api";
import type { UserDto } from "@repo/types";

export default function App() {
  const [users, setUsers] = useState<UserDto[]>([]);
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");

  async function refresh() {
    setUsers(await fetchUsers());
  }

  async function handleSignup(e: React.FormEvent) {
    e.preventDefault();
    await signup({ email, name, password: "changeme" });
    await refresh();
  }

  return (
    <main>
      <h1>Users</h1>
      <ul>{users.map((u) => <li key={u.id}>{u.email} — {u.name}</li>)}</ul>
      <form onSubmit={handleSignup}>
        <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" />
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="name" />
        <button>Sign up</button>
      </form>
    </main>
  );
}
