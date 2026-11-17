import { listMemberships, useListMemberships } from '@/api/gen/endpoints/identity/identity';

// Reads memberships page by page until next_cursor is null and answers as one page, so access is
// never decided from a partial list. The result sits under the generated list key.
async function listAllMemberships({ signal }: { signal: AbortSignal }) {
  const first = await listMemberships(undefined, { signal });
  const items = [...first.data.items];
  let cursor = first.data.next_cursor;
  while (cursor) {
    const page = await listMemberships({ cursor }, { signal });
    items.push(...page.data.items);
    cursor = page.data.next_cursor;
  }
  return { ...first, data: { items, next_cursor: null } };
}

export function useAllMemberships({ enabled = true }: { enabled?: boolean } = {}) {
  return useListMemberships(undefined, { query: { enabled, retry: false, queryFn: listAllMemberships } });
}
