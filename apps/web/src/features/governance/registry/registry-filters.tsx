import { useId } from 'react';

import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { uk } from '@/i18n/uk';

import type { FilterField, FilterValues } from './filters';

interface RegistryFiltersProps {
  values: FilterValues;
  onChange: (values: FilterValues) => void;
  onSubmit: () => void;
  // The fields the server refused with governance.search_invalid.
  invalid?: ReadonlySet<FilterField>;
  pending?: boolean;
}

// The search form of the реєстр: every field optional, combined with AND on the server.
export function RegistryFilters({ values, onChange, onSubmit, invalid = new Set(), pending = false }: RegistryFiltersProps) {
  const id = useId();
  const t = uk.registry.filters;
  const set = (field: FilterField) => (e: React.ChangeEvent<HTMLInputElement>) => onChange({ ...values, [field]: e.target.value });

  const field = (name: FilterField, label: string, message: string, input: (props: React.ComponentProps<'input'>) => React.ReactNode) => {
    const broken = invalid.has(name);
    return (
      <div className="space-y-1.5">
        <Label htmlFor={`${id}-${name}`}>{label}</Label>
        {input({ id: `${id}-${name}`, value: values[name], onChange: set(name), 'aria-invalid': broken || undefined, 'aria-describedby': broken ? `${id}-${name}-error` : undefined })}
        {broken && (
          <p id={`${id}-${name}-error`} className="text-[13px] text-destructive">
            {message}
          </p>
        )}
      </div>
    );
  };
  const text = (props: React.ComponentProps<'input'>) => <Input className="bg-card" {...props} />;
  const amount = (props: React.ComponentProps<'input'>) => <Input className="bg-card" type="number" inputMode="decimal" min={0} step="0.01" {...props} />;
  const date = (props: React.ComponentProps<'input'>) => <DateField className="bg-card" {...props} />;

  return (
    <form
      aria-label={t.label}
      className="space-y-4 rounded-lg border bg-card p-4"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {field('counterparty', t.counterparty, t.invalid.counterparty, text)}
        {field('subject', t.subject, t.invalid.subject, text)}
        {field('amountFrom', t.amountFrom, t.invalid.amount, amount)}
        {field('amountTo', t.amountTo, t.invalid.amount, amount)}
        {field('dateFrom', t.dateFrom, t.invalid.date, date)}
        {field('dateTo', t.dateTo, t.invalid.date, date)}
      </div>
      <Button type="submit" disabled={pending}>
        {t.submit}
      </Button>
    </form>
  );
}
