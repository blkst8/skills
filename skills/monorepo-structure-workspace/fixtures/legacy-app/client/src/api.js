const API_URL = import.meta.env.VITE_API_URL;

export async function fetchUsers() {
  const res = await fetch(`${API_URL}/api/users`);
  if (!res.ok) throw new Error("failed to fetch users");
  return res.json();
}

export async function signup(body: { email: string; name: string; password: string }) {
  const res = await fetch(`${API_URL}/api/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error("signup failed");
  return res.json();
}
