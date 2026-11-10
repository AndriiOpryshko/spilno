import { uk } from '@/i18n/uk';

// Says nothing about the record that was asked for: the same page for a missing one.
export function NoAccessPage() {
  return (
    <div className="mx-auto max-w-md px-4 py-24 text-center">
      <h1 className="text-2xl font-semibold">{uk.association.noAccess}</h1>
    </div>
  );
}
