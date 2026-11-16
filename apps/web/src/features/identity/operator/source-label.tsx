import { formatDate, uk } from '@/i18n/uk';

const sourceClass = 'text-xs text-muted-foreground';

export function ExtractSource({ checkedAt }: { checkedAt: string }) {
  return <span className={sourceClass}>{`${uk.operator.fromExtract} ${formatDate(checkedAt)}`}</span>;
}

export function SignatureSource({ checkedAt }: { checkedAt: string }) {
  return <span className={sourceClass}>{`${uk.operator.fromSignature} ${formatDate(checkedAt)}`}</span>;
}

export function NotChecked() {
  return <span className={sourceClass}>{uk.operator.notChecked}</span>;
}

export function EnteredByApplicant() {
  return <span className={sourceClass}>{uk.operator.enteredByApplicant}</span>;
}
