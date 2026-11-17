import { Navigate } from '@tanstack/react-router';

import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { InlineAlert } from '@/features/governance/inline-alert';
import { uk } from '@/i18n/uk';

import { isNoAccessProblem } from './entry';

// The one answer to a request that did not succeed (AC-05, AC-07): sign-in on 401, the no-access
// page on the access codes, and an alert that can read again for anything else (5xx, network).
export function RequestFallback({ error, onRetry, action }: { error: unknown; onRetry: () => void; action?: React.ReactNode }) {
  if ((error as { status?: number } | null)?.status === 401) return <Navigate to="/sign-in" replace />;
  if (isNoAccessProblem(error)) return <Navigate to="/no-access" replace />;
  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <InlineAlert
        title={uk.association.requestFailed}
        action={
          <>
            {action}
            <Button variant="outline" onClick={onRetry}>
              {uk.association.retry}
            </Button>
          </>
        }
      >
        {uk.association.requestFailedHint}
      </InlineAlert>
    </div>
  );
}

export function PageSkeleton() {
  return <Skeleton className="m-8 h-24 max-w-md" />;
}
