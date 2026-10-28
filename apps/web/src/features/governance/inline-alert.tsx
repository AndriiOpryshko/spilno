import { cn } from 'cn';

import { StatusGlyph } from './status-glyph';

interface InlineAlertProps {
  title: string;
  children?: React.ReactNode;
  // Next action, e.g. a retry or a link back.
  action?: React.ReactNode;
  // error: red title and glyph; neutral: an expected empty state.
  tone?: 'error' | 'neutral';
  className?: string;
}

// The role="alert" block (InlineAlert in Figma): load errors, a missing draft, a refused transition.
export function InlineAlert({ title, children, action, tone = 'error', className }: InlineAlertProps) {
  return (
    <div role="alert" className={cn('flex flex-wrap items-start gap-3 rounded-lg bg-destructive-subtle px-4 py-3', className)}>
      <StatusGlyph state={tone === 'error' ? 'blocked' : 'neutral'} className="mt-1" />
      <div className="min-w-0 flex-1">
        <p className={cn('font-semibold', tone === 'error' && 'text-destructive')}>{title}</p>
        {children && <p className="text-[13px]">{children}</p>}
      </div>
      {action}
    </div>
  );
}
