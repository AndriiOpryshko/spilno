import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import type { ApplicationReview } from '@/api/gen/model';

import { DecisionPanel } from './decision-panel';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';
const operatorEmail = 'operator@example.test';

const review: ApplicationReview = {
  application: {
    id: appId,
    kind: 'connection',
    edrpou: '12345678',
    association_id: null,
    association_name: 'ОСББ «Тестовий Будинок»',
    state: 'pending',
    filed_at: '2026-11-10T09:00:00Z',
    rejection_reason: null,
    created_at: '2026-11-10T09:00:00Z',
  } as ApplicationReview['application'],
  applicant_email: 'head@example.test',
  filings: [],
  checks: [],
  decisions: [],
  warnings: [],
  competing_application_ids: [],
  storage_account_email: null,
};

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });

// Replaces the global fetch for one story; the returned cleanup restores it.
const stubFetch = (refusal?: { status: number; code: string }) => () => {
  const original = globalThis.fetch;
  globalThis.fetch = (input, init) => {
    const path = new URL(String(input), 'http://localhost').pathname;
    if (path.endsWith('/identity/session')) {
      return Promise.resolve(json(200, { account_id: 'acc-1', email: operatorEmail, is_operator: true }));
    }
    if (init?.method === 'POST' && refusal) {
      return Promise.resolve(
        json(refusal.status, { type: 'about:blank', title: 'x', status: refusal.status, code: refusal.code }),
      );
    }
    return Promise.resolve(json(404, {}));
  };
  return () => {
    globalThis.fetch = original;
  };
};

const withQuery: Decorator = (Story) => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}>
    <Story />
  </QueryClientProvider>
);

const meta = {
  title: 'Identity/DecisionPanel',
  component: DecisionPanel,
  decorators: [withQuery],
  args: { applicationId: appId, filingId, review },
  beforeEach: stubFetch(),
} satisfies Meta<typeof DecisionPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ReadyToDecide: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: 'Схвалити' })).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: 'Відхилити' })).toBeInTheDocument();
    await expect(canvas.getByLabelText(/документ, на якому ґрунтується рішення/i)).toHaveAttribute('maxlength', '500');
  },
};

export const RejectDialogOpen: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'Відхилити' }));
    const body = within(canvasElement.ownerDocument.body);
    const dialog = await body.findByRole('dialog');
    await expect(within(dialog).getByLabelText(/Причина/)).toBeInTheDocument();
    await userEvent.click(within(dialog).getByRole('button', { name: /Відхилити/ }));
    await expect(await within(dialog).findByText('Вкажіть причину — її побачить заявник')).toBeInTheDocument();
  },
};

export const ServerRefusal: Story = {
  beforeEach: stubFetch({ status: 409, code: 'identity.signature_not_checked' }),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'Схвалити' }));
    await expect(
      await canvas.findByText('Спершу перевірте підпис КЕП і внесіть підписанта, керівника, код і дату витягу'),
    ).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: 'Схвалити' })).toBeInTheDocument();
  },
};

export const OwnApplication: Story = {
  args: { review: { ...review, applicant_email: operatorEmail } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByText('Це ваша власна заявка — оператор не ухвалює рішення щодо себе'),
    ).toBeInTheDocument();
    await waitFor(() => expect(canvas.queryByRole('button', { name: 'Схвалити' })).not.toBeInTheDocument());
    await expect(canvas.queryByRole('button', { name: 'Відхилити' })).not.toBeInTheDocument();
  },
};
