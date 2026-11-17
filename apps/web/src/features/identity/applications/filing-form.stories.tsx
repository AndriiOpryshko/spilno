import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createMemoryHistory, createRootRoute, createRouter } from '@tanstack/react-router';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import { expect as vitestExpect } from 'vitest';
import { page } from 'vitest/browser';

import { uk } from '@/i18n/uk';

import { FilingForm } from './filing-form';

const withQuery: Decorator = (Story) => (
  <QueryClientProvider client={new QueryClient()}>
    <Story />
  </QueryClientProvider>
);

const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const meta = {
  title: 'Identity/FilingForm',
  component: FilingForm,
  decorators: [withQuery, withRouter],
} satisfies Meta<typeof FilingForm>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Connection: Story = {
  args: { kind: 'connection' },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText('голова правління')).toBeInTheDocument();
    await document.fonts.ready;
    await vitestExpect(page.elementLocator(canvasElement)).toMatchScreenshot('filing-form');
  },
};

export const Transfer: Story = {
  args: { kind: 'transfer' },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByRole('button', { name: /Подати запит/ })).toBeInTheDocument();
  },
};

const problem = (status: number, code: string, detail: string, errors?: { pointer: string; detail: string }[]) =>
  new Response(JSON.stringify({ type: 'about:blank', title: code, status, code, detail, errors }), {
    status,
    headers: { 'Content-Type': 'application/problem+json' },
  });

// Replaces the global fetch for one story; the returned cleanup restores it.
const stubFetch = (response: () => Response) => () => {
  const original = globalThis.fetch;
  globalThis.fetch = () => Promise.resolve(response());
  return () => {
    globalThis.fetch = original;
  };
};

export const WithErrors: Story = {
  args: { kind: 'connection' },
  beforeEach: stubFetch(() =>
    problem(422, 'identity.application_invalid', 'The application is not valid.', [
      { pointer: '/edrpou', detail: 'must be 8 digits' },
      { pointer: '/term_start', detail: 'must be a date not later than the filing date' },
      { pointer: '/extract', detail: 'is required' },
    ]),
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: /Подати/ }));
    await waitFor(() => expect(canvas.getByLabelText(/ЄДРПОУ/)).toHaveAttribute('aria-invalid', 'true'));
    await expect(canvas.getByLabelText(/Початок строку/)).toHaveAttribute('aria-invalid', 'true');
    await expect(canvas.getByLabelText(/Витяг/)).toHaveAttribute('aria-invalid', 'true');
    await expect(canvas.getByRole('alert')).toHaveTextContent('Перевірте поля заявки: ЄДРПОУ, Початок строку, Витяг');
    await expect(canvas.getByText('ЄДРПОУ має містити 8 цифр')).toBeInTheDocument();
    await expect(canvas.getByText(uk.applications.filing.fieldErrors.extract as string)).toBeInTheDocument();
    await expect(canvas.queryByText(/must be|is required/)).not.toBeInTheDocument();
  },
};

const alreadyConnected =
  'Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу';

export const AlreadyConnected: Story = {
  args: { kind: 'connection' },
  beforeEach: stubFetch(() => problem(409, 'identity.association_already_connected', 'The application cannot be filed.')),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: /Подати/ }));
    await expect(await canvas.findByText(alreadyConnected)).toBeInTheDocument();
    const link = canvas.getByRole('link', { name: /запит на передачу/i });
    await expect(link.getAttribute('href')).toContain('kind=transfer');
  },
};
