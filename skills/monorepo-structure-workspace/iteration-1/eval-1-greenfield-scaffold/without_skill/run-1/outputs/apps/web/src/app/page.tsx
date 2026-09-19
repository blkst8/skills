import { fetchSummary } from '@/lib/api';

export const dynamic = 'force-dynamic';

export default async function HomePage() {
  // Placeholder until the first real feature lands: just proves the pipeline
  // (web -> @dashboard/shared contract -> api -> Postgres) end to end.
  let summary;
  try {
    summary = await fetchSummary();
  } catch {
    summary = { totalWidgets: 0, lastUpdated: new Date(0).toISOString() };
  }

  return (
    <main>
      <h1>Dashboard</h1>
      <p>{summary.totalWidgets} widgets, last updated {summary.lastUpdated}</p>
    </main>
  );
}
