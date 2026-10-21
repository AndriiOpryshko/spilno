import { createFileRoute } from '@tanstack/react-router';

import { HealthCard } from '@/features/system/health-card';

export const Route = createFileRoute('/')({
  component: HomePage,
});

function HomePage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Spilno</h1>
      <HealthCard />
    </div>
  );
}
