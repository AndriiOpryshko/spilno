import { listMemberships, useListMemberships } from '@/api/gen/endpoints/identity/identity';

const MAX_PAGES = 40;

// Reads memberships page by page until next_cursor is null and answers as one page, so access is
// never decided from a partial list. The result sits under the generated list key.
async function listAllMemberships({ signal }: { signal: AbortSignal }) {
  const first = await listMemberships(undefined, { signal });
  const items = [...first.data.items];
  let cursor = first.data.next_cursor;
  const seen = new Set<string>();
  while (cursor) {
    // A repeated cursor or a runaway chain is a broken server: fail so the layout shows its fallback.
    if (seen.has(cursor) || seen.size >= MAX_PAGES) throw new Error('memberships: pagination does not terminate');
    seen.add(cursor);
    const page = await listMemberships({ cursor }, { signal });
    items.push(...page.data.items);
    cursor = page.data.next_cursor;
  }
  return { ...first, data: { items, next_cursor: null } };
}

export function useAllMemberships({ enabled = true, refetchOnMount }: { enabled?: boolean; refetchOnMount?: boolean } = {}) {
  return useListMemberships(undefined, { query: { enabled, retry: false, refetchOnMount, queryFn: listAllMemberships } });
}
