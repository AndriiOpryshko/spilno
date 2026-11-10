import { expect, type Browser, type Page } from '@playwright/test';

// Helpers shared by the association-onboarding specs. Every spec makes its own people and its own
// ЄДРПОУ (random), so the specs run in parallel without sharing a code or an account.

export const operatorEmail = process.env.E2E_OPERATOR_EMAIL ?? 'operator@example.test';

const digits = () => String(Math.floor(10_000_000 + Math.random() * 89_999_999));
const token = () => Math.random().toString(36).slice(2, 10);

export const newCode = digits;
export const newEmail = (role: string) => `${role}-${token()}@example.test`;

export interface Filing {
  code: string;
  name: string;
  address: string;
  fullName: string;
  termStart: string;
  termEnd?: string;
}

export const newFiling = (): Filing => {
  const code = newCode();
  return {
    code,
    name: `ОСББ Тест ${code}`,
    address: 'вул. Тестова, 1',
    fullName: 'Тест Тестенко Тестович',
    termStart: '2025-01-01',
  };
};

export const daysAgo = (n: number) => new Date(Date.now() - n * 86_400_000).toISOString().slice(0, 10);
export const today = () => new Date().toLocaleDateString('uk-UA');

export const extractFile = { name: 'extract.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test extract') };

// The sign-in shim (SPILNO_DEV_SIGN_IN) takes the code as a verified email address: start the
// sign-in for the state, hand the callback the address, keep the session cookie in this browser.
export async function signInAs(page: Page, email: string) {
  const api = page.context().request;
  const start = await api.get('/api/identity/sign-in', { maxRedirects: 0 });
  const location = start.headers()['location'];
  expect(location, 'sign-in start must redirect to the provider').toBeTruthy();
  const state = new URL(location ?? '', 'http://localhost').searchParams.get('state') ?? '';
  const done = await api.get('/api/identity/sign-in/callback', { params: { code: email, state }, maxRedirects: 0 });
  expect(done.status(), 'sign-in callback must set the session').toBe(302);
}

export async function sessionFor(browser: Browser, email: string) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await signInAs(page, email);
  return page;
}

export async function fillFiling(page: Page, f: Partial<Filing>) {
  const text = (label: string, value: string | undefined) =>
    value === undefined ? undefined : page.getByLabel(label, { exact: true }).fill(value);
  await text('ЄДРПОУ', f.code);
  await text('Назва ОСББ', f.name);
  await text('Адреса', f.address);
  await text('ПІБ голови', f.fullName);
  await text('Початок строку повноважень', f.termStart);
  await text('Кінець строку повноважень (необов’язково)', f.termEnd);
}

export async function fileConnection(page: Page, f: Filing) {
  await page.goto('/applications/new?kind=connection');
  await fillFiling(page, f);
  await page.getByLabel('Витяг (до 5 МБ)').setInputFiles(extractFile);
  await page.getByRole('button', { name: 'Подати заявку' }).click();
}

export async function fileAndConfirm(page: Page, f: Filing) {
  await fileConnection(page, f);
  await expect(page.getByText('Заявку подано, очікує розгляду оператором')).toBeVisible();
}

// The id of the applicant's заявка, read from their own list.
export async function applicationIdOf(page: Page, name: string) {
  await page.goto('/applications');
  const href = await page.getByRole('link', { name, exact: true }).getAttribute('href');
  return (href ?? '').split('/').pop() ?? '';
}

export interface Check {
  signer: string;
  head: string;
  edrpou: string;
  datedOn: string;
}

export const checkFor = (f: Filing, over: Partial<Check> = {}): Check => ({
  signer: f.fullName,
  head: f.fullName,
  edrpou: f.code,
  datedOn: daysAgo(1),
  ...over,
});

export async function openReview(page: Page, applicationId: string) {
  await page.goto(`/operator/applications/${applicationId}`);
  await expect(page.getByRole('heading', { name: 'Перевірка витягу' })).toBeVisible();
}

export async function recordCheck(page: Page, c: Check) {
  await page.getByLabel('Підписант', { exact: true }).fill(c.signer);
  await page.getByLabel('Дата підпису').fill(daysAgo(1));
  await page.getByLabel('Керівник за витягом').fill(c.head);
  await page.getByLabel('ЄДРПОУ за витягом').fill(c.edrpou);
  await page.getByLabel('Дата витягу').fill(c.datedOn);
  await page.getByRole('button', { name: 'Записати перевірку' }).click();
  await expect(page.getByText(/^з КЕП, перевірено оператором/).first()).toBeVisible();
}

// A connected association through the UI: file, check, approve. Returns the applicant's page.
export async function connectAssociation(browser: Browser, f: Filing, applicantEmail: string) {
  const applicant = await sessionFor(browser, applicantEmail);
  await fileAndConfirm(applicant, f);
  const id = await applicationIdOf(applicant, f.name);
  const operator = await sessionFor(browser, operatorEmail);
  await openReview(operator, id);
  await recordCheck(operator, checkFor(f));
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();
  await operator.context().close();
  return { applicant, id };
}
