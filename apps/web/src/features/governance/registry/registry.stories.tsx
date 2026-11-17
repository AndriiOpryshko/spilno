import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { getDownloadFileUrl } from '@/api/gen/endpoints/governance/governance';
import type { FileRef, SignedProtocol } from '@/api/gen/model';

import { emptyFilters } from './filters';
import { RegistryFilters } from './registry-filters';
import { pageRows, RegistryResults } from './registry-results';
import { EmptyRegistry } from './registry-screen';
import { SignedProtocolView } from './signed-protocol-view';

const association = '00000000-0000-7000-8000-000000000001';
const member = '00000000-0000-7000-8000-00000000b001';

// Links need a router; the story itself is the root route's component.
const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const decision = (n: number, subject: string, counterparty: string | null, amount: number | null) => ({
  protocol_id: `00000000-0000-7000-8000-000000000e${n}`,
  agenda_item_id: `00000000-0000-7000-8000-0000000010${n}`,
  number: n,
  meeting_date: '2026-11-05',
  subject,
  counterparty,
  amount_kopiyky: amount,
  currency: amount === null ? null : ('UAH' as const),
});

const rows = pageRows(
  [decision(14, 'Про затвердження договору на гідроізоляцію покрівлі', 'ТОВ Тест Покрівля', 18450000), decision(12, 'Про заміну лічильників води', 'ФОП Тест Лічильник', 4200000)],
  [13],
);

const resultsMeta = {
  title: 'Governance/Registry',
  component: RegistryResults,
  decorators: [withRouter],
  args: { associationId: association, rows, hasMore: true, onMore: () => {} },
} satisfies Meta<typeof RegistryResults>;

export default resultsMeta;
type Story = StoryObj<typeof resultsMeta>;

export const Results: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('link', { name: /гідроізоляцію/ })).toBeVisible();
    await expect(canvas.getByText(/№ 13 — знято з розгляду/)).toBeVisible();
    await expect(canvas.getByRole('button', { name: 'Показати ще' })).toBeVisible();
  },
};

export const LastPage: Story = {
  args: { hasMore: false },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('button')).not.toBeInTheDocument();
  },
};

export const Empty: Story = {
  render: () => <EmptyRegistry filtered={false} />,
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/немає підписаних протоколів/)).toBeVisible();
  },
};

export const NothingFound: Story = {
  render: () => <EmptyRegistry filtered />,
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('Нічого не знайдено')).toBeVisible();
  },
};

export const RefusedFilters: Story = {
  render: () => <RegistryFilters values={{ ...emptyFilters, amountFrom: '3000', amountTo: '10' }} onChange={() => {}} onSubmit={() => {}} invalid={new Set(['amountFrom', 'amountTo'])} />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const from = canvas.getByLabelText(/Сума від/);
    await expect(from).toHaveAttribute('aria-invalid', 'true');
    await expect(from).toHaveAccessibleDescription(/\S/);
    await expect(canvas.getByLabelText(/Контрагент/)).not.toHaveAttribute('aria-invalid');
  },
};

const file = (id: string, name: string): FileRef => ({
  id,
  original_name: name,
  media_type: 'application/pdf',
  size_bytes: 91234,
  sha256: '2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
  uploaded_by_member_id: member,
  created_at: '2026-11-01T12:00:04Z',
});

const itemId = '00000000-0000-7000-8000-000000001001';
const signed: SignedProtocol = {
  protocol_id: '00000000-0000-7000-8000-000000000e14',
  number: 14,
  meeting_date: '2026-11-05',
  meeting: { id: '00000000-0000-7000-8000-000000000f01', held_at: '2026-11-05T16:00:00Z', meeting_date: '2026-11-05', chair_board_member_id: member },
  agenda_items: [
    {
      id: itemId,
      position: 1,
      case_id: null,
      subject: 'Про затвердження договору на гідроізоляцію покрівлі',
      fragments: [{ text: 'Затвердити договір на гідроізоляцію покрівлі з ТОВ Тест Покрівля.' }],
      wording_author_member_id: member,
      wording_revised_at: '2026-11-01T12:00:00Z',
      slots: [],
    },
  ],
  documents: [
    {
      id: '00000000-0000-7000-8000-000000000d02',
      case_id: '00000000-0000-7000-8000-000000000c01',
      file: file('00000000-0000-7000-8000-000000000702', 'rakhunok-123.pdf'),
      previous_version_id: null,
      document_date: '2026-10-18',
      created_by_member_id: member,
      created_at: '2026-10-20T09:10:00Z',
    },
  ],
  print_copy: {
    id: '00000000-0000-7000-8000-000000001301',
    protocol_revision: 9,
    number: 14,
    file: file('00000000-0000-7000-8000-000000000703', 'protocol-14.pdf'),
    requested_by_member_id: member,
    created_at: '2026-11-01T12:00:04Z',
    voided_at: null,
  },
  meeting_facts: {
    signed_scan: file('00000000-0000-7000-8000-000000000704', 'protocol-14-signed.pdf'),
    present_member_ids: [member],
    item_votes: [{ agenda_item_id: itemId, votes_for: 3, votes_against: 0, votes_abstained: 0 }],
    has_handwritten_corrections: true,
    corrections_note: 'У пункті 1 виправлено номер рахунку від руки, підписано всіма присутніми',
    entered_by_member_id: member,
    entered_at: '2026-11-06T08:00:00Z',
  },
  poll_history: [
    {
      id: '00000000-0000-7000-8000-000000001401',
      round_number: 1,
      protocol_revision: 8,
      label: 'Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його',
      opened_by_member_id: member,
      created_at: '2026-10-27T11:00:00Z',
      closed_at: '2026-10-27T12:00:00Z',
      closed_by_member_id: member,
      frozen_agenda_items: [],
      items: [{ agenda_item_id: itemId, tally: { for: 3, against: 1, abstain: 1 }, votes: [] }],
    },
  ],
  locked_by_member_id: member,
  locked_at: '2026-11-06T08:00:00Z',
};

export const SignedView: Story = {
  render: () => <SignedProtocolView associationId={association} protocol={signed} />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('heading', { level: 1, name: /14/ })).toBeVisible();
    await expect(canvas.getByRole('link', { name: /protocol-14-signed\.pdf/ })).toHaveAttribute('href', getDownloadFileUrl(association, '00000000-0000-7000-8000-000000000704'));
    await expect(canvas.getByText(/не є голосуванням засідання/)).toBeVisible();
    await expect(canvas.queryByRole('textbox')).not.toBeInTheDocument();
    await expect(canvas.queryByRole('button')).not.toBeInTheDocument();
  },
};
