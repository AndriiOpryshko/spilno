import { useGetHealth } from '@/api/gen/endpoints/system/system';
import type { Health } from '@/api/gen/model';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function HealthCard() {
  const query = useGetHealth({ query: { retry: false } });
  // A 503 still carries a Health body; the generated client puts it in error.info.
  const health = query.data?.data ?? query.error?.info;
  return <HealthView health={health} loading={query.isPending} />;
}

export function HealthView({ health, loading }: { health: Health | undefined; loading: boolean }) {
  return (
    <Card className="max-w-sm">
      <CardHeader>
        <CardTitle>Service health</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {loading ? (
          <p className="text-muted-foreground">Checking...</p>
        ) : !health ? (
          <p role="alert">Core service is unreachable.</p>
        ) : (
          <>
            <Row label="Status">
              <Badge variant={health.status === 'ok' ? 'default' : 'destructive'}>{health.status}</Badge>
            </Row>
            <Row label="Database">
              <span>
                {health.database.status}
                {health.database.version ? ` (Postgres ${health.database.version})` : ''}
              </span>
            </Row>
          </>
        )}
      </CardContent>
    </Card>
  );
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      {children}
    </div>
  );
}
