import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import type { BoardMember } from '@/api/gen/model';

import { RosterScreen } from './roster-screen';

const associationId = '00000000-0000-7000-8000-000000000002';
const decisionId = '00000000-0000-7000-8000-00000000f002';
const extractId = '00000000-0000-7000-8000-00000000e001';

const entry = (over: Partial<BoardMember>): BoardMember => ({
  id: '00000000-0000-7000-8000-00000000b001',
  full_name: 'Тестовий Голова Правління',
  sign_in_email: 'head@example.test',
  term_start: '2025-03-01',
  term_end: null,
  revised_by_member_id: null,
  revised_at: '2025-03-01T08:00:00Z',
  position: 'head',
  decision_id: null,
  ...over,
});
const filled: BoardMember[] = [
  entry({ id: '00000000-0000-7000-8000-00000000b002', full_name: 'Новий Голова Тестовий', term_start: '2026-10-25', decision_id: decisionId }),
  entry({ id: '00000000-0000-7000-8000-00000000b001', full_name: 'Колишній Голова Тестовий', term_start: '2024-01-15', term_end: '2026-10-24', decision_id: decisionId }),
  entry({ id: '00000000-0000-7000-8000-00000000b003', full_name: 'Звичайний Член Тестовий', position: 'member' }),
];
const card = {
  id: associationId,
  name: 'ОСББ «Тестовий Будинок»',
  edrpou: '12345678',
  address: 'м. Київ, вул. Тестова, 1',
  state: 'connected',
  head_full_name: 'Новий Голова Тестовий',
  connected_by_hand: false,
  history: [
    {
      kind: 'transfer_approved',
      occurred_at: '2026-10-21T08:05:00Z',
      decision_id: decisionId,
      filed_at: '2026-10-20T09:00:00Z',
      reason: null,
      head_full_name: 'Новий Голова Тестовий',
      ended_head_full_name: 'Колишній Голова Тестовий',
      signer_full_name: 'Новий Голова Тестовий',
      extract_id: extractId,
      extract_dated_on: '2026-10-18',
    },
  ],
};

const meta = { title: 'Governance/RosterScreen', component: RosterScreen, args: { associationId } } satisfies Meta<typeof RosterScreen>;

export default meta;
type Story = StoryObj<typeof meta>;

// Replaces the global fetch for one story: the roster and the card answer like the contract, or the roster read fails.
const stubApi = (items: BoardMember[] | 'fail') => () => {
  const original = globalThis.fetch;
  globalThis.fetch = (input: RequestInfo | URL) => {
    const path = new URL(String(input), 'http://localhost').pathname;
    const ok = (body: unknown) => Promise.resolve(new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    if (items === 'fail') return Promise.resolve(new Response('boom', { status: 500 }));
    if (path.endsWith('/board-members')) return ok({ items, next_cursor: null });
    if (path.endsWith('/identity/memberships')) return ok({ items: [], next_cursor: null });
    return ok(card);
  };
  return () => {
    globalThis.fetch = original;
  };
};

export const Empty: Story = {
  beforeEach: stubApi([]),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText('Правління ще не заповнене')).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: 'Додати члена правління' })).toBeInTheDocument();
  },
};

export const Filled: Story = {
  beforeEach: stubApi(filled),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText('Новий Голова Тестовий')).toBeInTheDocument();
    await expect(await canvas.findByText(/призначено за запитом на передачу від/)).toBeInTheDocument();
    await expect(await canvas.findByText(/повноваження завершено за запитом на передачу від/)).toBeInTheDocument();
    await expect(canvas.getAllByRole('link', { name: 'Витяг' })).toHaveLength(2);
    await expect(canvas.getByText('Звичайний Член Тестовий')).toBeInTheDocument();
  },
};

export const AddForm: Story = {
  beforeEach: stubApi(filled),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: 'Додати члена правління' }));
    const body = within(canvasElement.ownerDocument.body);
    await expect(await body.findByLabelText(/ПІБ/)).toBeInTheDocument();
    await expect(body.getByLabelText(/Початок строку/)).toBeInTheDocument();
  },
};

export const LoadError: Story = {
  beforeEach: stubApi('fail'),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('alert')).toHaveTextContent('Не вдалося виконати запит');
    await expect(canvas.getByRole('button', { name: 'Спробувати ще раз' })).toBeInTheDocument();
  },
};
