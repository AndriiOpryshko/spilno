import { uk } from '@/i18n/uk';

// 2026-11-11T08:05:00Z -> 11.11.2026
export const formatDay = (iso: string) => {
  const [y, m, d] = iso.slice(0, 10).split('-');
  return `${d}.${m}.${y}`;
};

const sourceClass = 'text-xs text-muted-foreground';

export function ExtractSource({ checkedAt }: { checkedAt: string }) {
  return <span className={sourceClass}>{`${uk.operator.fromExtract} ${formatDay(checkedAt)}`}</span>;
}

export function SignatureSource({ checkedAt }: { checkedAt: string }) {
  return <span className={sourceClass}>{`${uk.operator.fromSignature} ${formatDay(checkedAt)}`}</span>;
}

export function NotChecked() {
  return <span className={sourceClass}>{uk.operator.notChecked}</span>;
}
