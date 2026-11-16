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
          <div key={m.association_id} className="flex gap-2">
            <Link to="/o/$associationId/cases" params={{ associationId: m.association_id }} className="flex-1 rounded-md border px-3 py-2 hover:bg-accent">
              {m.association_name}
            </Link>
            <Link to="/o/$associationId/association" params={{ associationId: m.association_id }} className="rounded-md border px-3 py-2 hover:bg-accent">
              {uk.association.openCard}
            </Link>
          </div>
        ))}
        {isOperator && (
          <Link to="/operator" className="rounded-md border px-3 py-2 hover:bg-accent">
            {uk.operator.entry}
          </Link>
        )}
      </CardContent>
    </Card>
  );
}
