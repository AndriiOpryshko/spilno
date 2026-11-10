import { createFileRoute } from '@tanstack/react-router';

import { HealthCard } from '@/features/system/health-card';

// Public: the entry at / needs a session, the service health does not.
export const Route = createFileRoute('/status')({
  component: StatusPage,
});

function StatusPage() {
  return (
    <div className="space-y-6 px-4 py-8 lg:px-8">
      <h1 className="text-2xl font-semibold">Spilno</h1>
      <HealthCard />
    </div>
  );
}
