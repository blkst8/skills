import type { User } from "../../lib/apiClient";

export function UsersList({ users }: { users: User[] }) {
  if (users.length === 0) {
    return <p>No users yet.</p>;
  }
  return (
    <ul>
      {users.map((u) => (
        <li key={u.id}>
          {u.name} ({u.email})
        </li>
      ))}
    </ul>
  );
}
