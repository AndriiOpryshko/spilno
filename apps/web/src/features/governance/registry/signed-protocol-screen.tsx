import { useQuery } from '@tanstack/react-query';
import { useParams } from '@tanstack/react-router';

import { getSignedProtocol } from '@/api/gen/endpoints/governance/governance';
import { listMemberships } from '@/api/gen/endpoints/identity/identity';
import { TopBar } from '@/components/app-shell';
import { Skeleton } from '@/components/ui/skeleton';
import { uk } from '@/i18n/uk';

import { InlineAlert } from '../inline-alert';
import { SignedProtocolView } from './signed-protocol-view';

// One signed protocol of the реєстр, read only (AC-26, AC-09; board-protocol T43).
export function SignedProtocolScreen({ protocolId }: { protocolId: string }) {
  const { associationId } = useParams({ from: '/o/$associationId' });
  const query = useQuery({ queryKey: ['signed-protocol', associationId, protocolId], queryFn: () => getSignedProtocol(associationId, protocolId), retry: false });
  const protocol = query.data?.data;
  // The layout already loaded the roster; this reads the same cache entry.
  const { data: roster } = useQuery({ queryKey: ['memberships'], queryFn: () => listMemberships() });
  const associationName = roster?.data.items.find((m) => m.association_id === associationId)?.association_name;

  const crumbs = [{ label: uk.nav.registry, to: `/o/${associationId}/registry` }, ...(protocol ? [{ label: uk.registry.signed.breadcrumb(protocol.number) }] : [])];
  return (
    <>
      <TopBar crumbs={crumbs} associationName={associationName} />
      <div className="px-4 py-6 lg:px-8 lg:py-8">
        {query.isPending && (
          <div className="space-y-2" aria-busy>
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-32 w-full rounded-lg" />
          </div>
        )}
        {query.isError && <InlineAlert title={uk.registry.signed.loadError}>{uk.registry.signed.loadErrorHint}</InlineAlert>}
        {protocol && <SignedProtocolView associationId={associationId} protocol={protocol} />}
      </div>
    </>
  );
}
