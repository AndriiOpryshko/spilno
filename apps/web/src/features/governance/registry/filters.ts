import type { SearchRegistryParams } from '@/api/gen/model';

// The registry filter form as typed: amounts in hryvnias, dates as YYYY-MM-DD.
export interface FilterValues {
  counterparty: string;
  subject: string;
  amountFrom: string;
  amountTo: string;
  dateFrom: string;
  dateTo: string;
}

export const emptyFilters: FilterValues = { counterparty: '', subject: '', amountFrom: '', amountTo: '', dateFrom: '', dateTo: '' };

export type FilterField = keyof FilterValues;

const toKopiyky = (hryvnias: string) => {
  const value = Number(hryvnias.replace(',', '.'));
  return Number.isFinite(value) ? Math.round(value * 100) : undefined;
};

// Only the filled fields go on the wire, under the contract's names.
export function toSearchParams(values: FilterValues): SearchRegistryParams {
  const text = (value: string) => value.trim() || undefined;
  const amount = (value: string) => (value.trim() ? toKopiyky(value.trim()) : undefined);
  return {
    counterparty: text(values.counterparty),
    subject: text(values.subject),
    amount_from_kopiyky: amount(values.amountFrom),
    amount_to_kopiyky: amount(values.amountTo),
    date_from: text(values.dateFrom),
    date_to: text(values.dateTo),
  };
}

export const isFiltered = (values: FilterValues) => Object.values(toSearchParams(values)).some((v) => v !== undefined);

const fieldOfParam: Record<string, FilterField> = {
  counterparty: 'counterparty',
  subject: 'subject',
  amount_from_kopiyky: 'amountFrom',
  amount_to_kopiyky: 'amountTo',
  date_from: 'dateFrom',
  date_to: 'dateTo',
};

type SearchProblem = { status?: number; info?: { code?: string; errors?: { pointer: string }[] } };

// governance.search_invalid names the broken query parameters; the form fields they belong to.
export function invalidFields(error: unknown): { invalid: boolean; fields: Set<FilterField> } {
  const problem = (error ?? {}) as SearchProblem;
  const invalid = problem.status === 422 && problem.info?.code === 'governance.search_invalid';
  const fields = new Set<FilterField>();
  if (invalid) {
    for (const { pointer } of problem.info?.errors ?? []) {
      const field = fieldOfParam[pointer.split('/').pop() ?? ''];
      if (field) fields.add(field);
    }
  }
  return { invalid, fields };
}
