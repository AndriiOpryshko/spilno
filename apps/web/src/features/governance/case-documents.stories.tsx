import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { getDownloadFileUrl } from '@/api/gen/endpoints/governance/governance';
import type { CaseDocument } from '@/api/gen/model';

import { CaseDocuments } from './case-documents';

const association = '00000000-0000-7000-8000-000000000001';
const caseId = '00000000-0000-7000-8000-000000000c01';
const member = '00000000-0000-7000-8000-00000000b001';

// Links need a router; the story itself is the root route's component.
const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const document = (id: string, fileId: string, name: string, previous: string | null, date: string): CaseDocument => ({
  id,
  case_id: caseId,
  file: {
    id: fileId,
    original_name: name,
    media_type: 'application/pdf',
    size_bytes: 482113,
    sha256: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
    uploaded_by_member_id: member,
    created_at: '2026-10-20T09:10:00Z',
  },
  previous_version_id: previous,
  document_date: date,
  created_by_member_id: member,
  created_at: '2026-10-20T09:10:00Z',
});

const first = document('00000000-0000-7000-8000-000000000d01', '00000000-0000-7000-8000-000000000701', 'kosztorys-v1.pdf', null, '2026-10-01');
const second = document('00000000-0000-7000-8000-000000000d02', '00000000-0000-7000-8000-000000000702', 'kosztorys-v2.pdf', first.id, '2026-10-18');

// Contract for the implementer: CaseDocuments({ associationId, caseId, documents }) lists the documents with their
// versions, links every file through getDownloadFileUrl, and holds the upload (a «Файл…» FileField and its button).
// An empty list is explained, not left bare.
const meta = {
  title: 'Governance/CaseDocuments',
  component: CaseDocuments,
  decorators: [withRouter],
  args: { associationId: association, caseId, documents: [first, second] },
} satisfies Meta<typeof CaseDocuments>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithVersions: Story = {
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('kosztorys-v1.pdf')).toBeVisible();
    await expect(canvas.getByText('kosztorys-v2.pdf')).toBeVisible();
    const hrefs = canvas.getAllByRole('link').map((a) => a.getAttribute('href'));
    await expect(hrefs).toContain(getDownloadFileUrl(association, second.file.id));
    await expect(canvas.getByLabelText(/Файл/)).toBeVisible();
    await expect(canvas.getByRole('button', { name: /Додати|Завантажити|Прикріпити/ })).toBeVisible();
  },
};

export const NoDocuments: Story = {
  args: { documents: [] },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText(/ще немає документів/i)).toBeVisible();
    await expect(canvas.getByLabelText(/Файл/)).toBeVisible();
  },
};
