import type { Membership, Principal } from '@/api/gen/model';

export type Entry = { kind: 'sign-in' } | { kind: 'applications' } | { kind: 'association'; associationId: string } | { kind: 'picker' };

// Where `/` sends a person (AC-04d): the choice of association lives in the address, not the session.
export function resolveEntry({ session, memberships }: { session: Principal | undefined; memberships: Membership[] }): Entry {
  if (!session) return { kind: 'sign-in' };
  const [only] = memberships;
  if (memberships.length === 1 && only) return { kind: 'association', associationId: only.association_id };
  if (memberships.length === 0 && !session.is_operator) return { kind: 'applications' };
  return { kind: 'picker' };
}

const noAccessCodes = ['identity.not_current_member', 'identity.not_operator'];

// A 403 that answers the same for an existing and a missing record (AC-07, AC-16, AC-16b).
export function isNoAccessProblem(err: unknown): boolean {
  const { status, info } = (err ?? {}) as { status?: number; info?: { code?: string } };
  if (status !== 403) return false;
  const code = info?.code;
  return code !== undefined && noAccessCodes.includes(code);
}
