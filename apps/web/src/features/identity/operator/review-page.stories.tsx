import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { expect as vitestExpect } from 'vitest';
import { page } from 'vitest/browser';

import type { ApplicationReview } from '@/api/gen/model';

import { ReviewPage } from './review-page';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';

const base: ApplicationReview = {
  application: {
    id: appId,
    kind: 'connection',
    edrpou: '12345678',
    association_id: null,
    association_name: 'ОСББ «Липки»',
    state: 'pending',
    filed_at: '2026-11-10T09:00:00Z',
    rejection_reason: null,
    created_at: '2026-11-10T09:00:00Z',
  } as ApplicationReview['application'],
  applicant_email: 'head@example.test',
  filings: [
    {
      id: filingId,
      association_name: 'ОСББ «Липки»',
      association_address: 'м. Київ, вул. Тестова, 1',
      applicant_full_name: 'Тестовий Голова Правління',
      position: 'head',
      term_start: '2026-01-01',
      term_end: null,
      extract: { id: 'e1', original_name: 'vytyah.pdf', media_type: 'application/pdf', size_bytes: 10, sha256: 'a'.repeat(64) },
      created_at: '2026-11-10T09:00:00Z',
    },
  ],
  checks: [],
  decisions: [],
  warnings: [],
  competing_application_ids: [],
  storage_account_email: null,
};

const check = {
  id: 'c1',
  filing_id: filingId,
  signature_valid: true,
  signer_full_name: 'Тестовий Голова Правління',
  signed_on: '2026-11-08',
  head_full_name: 'Тестовий Голова Правління',
  extract_edrpou: '12345678',
  extract_dated_on: '2026-11-05',
  checked_at: '2026-11-11T08:05:00Z',
};

const serve = (review: ApplicationReview) => () => {
  globalThis.fetch = () =>
    Promise.resolve(new Response(JSON.stringify(review), { status: 200, headers: { 'Content-Type': 'application/json' } }));
};

const meta = {
  title: 'Identity/ReviewPage',
  component: ReviewPage,
  args: { applicationId: appId },
  decorators: [
    (Story) => (
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <Story />
      </QueryClientProvider>
    ),
  ],
} satisfies Meta<typeof ReviewPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Unchecked: Story = {
  beforeEach: serve(base),
  play: async ({ canvasElement }) => {
    await expect((await within(canvasElement).findAllByText(/не перевірено/))[0]).toBeInTheDocument();
  },
};

export const Clean: Story = {
  beforeEach: serve({ ...base, checks: [check] }),
  play: async ({ canvasElement }) => {
    await expect((await within(canvasElement).findAllByText(/з КЕП, перевірено оператором/))[0]).toBeInTheDocument();
    await document.fonts.ready;
    await vitestExpect(page.elementLocator(canvasElement)).toMatchScreenshot('review-page');
  },
};

export const Mismatch: Story = {
  beforeEach: serve({ ...base, checks: [check], warnings: [{ kind: 'name_mismatch', detail: 'Іваненко І. І. ≠ Петренко П. П.' }] }),
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/Іваненко/)).toBeInTheDocument();
  },
};

export const OldExtract: Story = {
  beforeEach: serve({ ...base, checks: [check], warnings: [{ kind: 'extract_age', detail: '2026-09-01' }] }),
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/старший за 30 днів/)).toBeInTheDocument();
  },
};

export const Competing: Story = {
  beforeEach: serve({ ...base, competing_application_ids: ['00000000-0000-7000-8000-0000000000a2'] }),
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/конкуруюча заявка/)).toBeInTheDocument();
  },
};

const refiledId = '00000000-0000-7000-8000-0000000000f2';

// A refiled заявка: filings newest first, checks oldest first, as the server returns them.
export const Refiled: Story = {
  beforeEach: serve({
    ...base,
    filings: [
      { ...base.filings[0]!, id: refiledId, applicant_full_name: 'Новий Заявник', created_at: '2026-11-12T09:00:00Z' },
      { ...base.filings[0]!, applicant_full_name: 'Попередній Заявник' },
    ],
    checks: [
      { ...check, signer_full_name: 'Підписант першої заяви' },
      { ...check, id: 'c2', filing_id: refiledId, signer_full_name: 'Підписант нової заяви', checked_at: '2026-11-13T09:00:00Z' },
    ],
  }),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText('Новий Заявник')).toBeInTheDocument();
    await expect(canvas.getByText('Підписант нової заяви')).toBeInTheDocument();
    await expect(canvas.getByText('Попередній Заявник')).toBeInTheDocument();
    await expect(canvas.getAllByRole('link', { name: /витяг/i })).toHaveLength(2);
    await expect(canvas.getAllByRole('button', { name: /Записати перевірку/ })).toHaveLength(1);
  },
};
