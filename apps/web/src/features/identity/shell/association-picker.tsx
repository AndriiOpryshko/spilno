import { Link } from '@tanstack/react-router';

import type { Membership } from '@/api/gen/model';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { uk } from '@/i18n/uk';

export function AssociationPicker({ memberships, isOperator }: { memberships: Membership[]; isOperator: boolean }) {
  return (
    <Card className="mx-auto mt-16 max-w-md">
      <CardHeader>
        <CardTitle>{uk.association.pickTitle}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-2 text-sm">
        {memberships.map((m) => (
          <Link key={m.association_id} to="/o/$associationId/cases" params={{ associationId: m.association_id }} className="rounded-md border px-3 py-2 hover:bg-accent">
            {m.association_name}
          </Link>
        ))}
        {isOperator && (
          <a href="/operator/" className="rounded-md border px-3 py-2 hover:bg-accent">
            {uk.operator.entry}
          </a>
        )}
      </CardContent>
    </Card>
  );
}
