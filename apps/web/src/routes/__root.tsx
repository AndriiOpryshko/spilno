import { type QueryClient, useQuery } from '@tanstack/react-query';
import { createRootRouteWithContext, Outlet } from '@tanstack/react-router';

import { Sidebar } from '@/components/app-shell';
import { listCases } from '@/features/governance/api';

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: RootLayout,
});

function RootLayout() {
  const cases = useQuery({ queryKey: ['cases'], queryFn: listCases });
  return (
    <div className="flex min-h-svh">
      <Sidebar casesCount={cases.data?.length} />
      <main className="min-w-0 flex-1">
        <Outlet />
      </main>
    </div>
  );
}
