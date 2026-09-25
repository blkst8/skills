import { useEffect, useState } from "react";
import { UsersList } from "./features/users/UsersList";
import { api, type User } from "./lib/apiClient";

export default function App() {
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    api
      .listUsers()
      .then(setUsers)
      .catch((err) => console.error("failed to load users", err));
  }, []);

  return (
    <main>
      <h1>Side Project</h1>
      <UsersList users={users} />
    </main>
  );
}
