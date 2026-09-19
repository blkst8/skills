import { useEffect, useState } from "react";
import { fetchHealth, type HealthStatus } from "./api/client";

export default function App() {
  const [status, setStatus] = useState<HealthStatus | null>(null);

  useEffect(() => {
    fetchHealth().then(setStatus).catch(() => setStatus(null));
  }, []);

  return (
    <main>
      <h1>Sideproject</h1>
      <p>API status: {status ? status.status : "unreachable"}</p>
    </main>
  );
}
