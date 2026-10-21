import { Link } from '@tanstack/react-router';
import { cn } from 'cn';

import { uk } from '@/i18n/uk';

export function BrandMark({ className }: { className?: string }) {
  return (
    <span className={cn('inline-flex size-7 shrink-0 items-center justify-center rounded-md bg-linear-135 from-brand-from via-brand-via via-30% to-brand-to text-base font-bold text-white', className)}>
      с
    </span>
  );
}

const soon = [uk.nav.protocols, uk.nav.meetings, uk.nav.documents];

// Desktop navigation (Sidebar in Figma). Hidden below lg; screens render their own mobile top bar.
export function Sidebar({ casesCount }: { casesCount?: number }) {
  return (
    <aside className="sticky top-0 hidden h-svh w-60 shrink-0 flex-col gap-1 bg-sidebar px-3 py-4 text-sidebar-foreground lg:flex">
      <div className="flex items-center gap-3 p-2">
        <BrandMark />
        <div className="min-w-0">
          <p className="text-sm font-semibold">{uk.app.name}</p>
          <p className="truncate text-[11px] opacity-65">{uk.app.osbb}</p>
        </div>
      </div>
      <div className="flex items-center justify-between rounded-md bg-sidebar-accent px-3 py-2 text-[13px] opacity-65" aria-hidden>
        <span>{uk.app.search}</span>
        <span className="text-[11px] font-medium">⌘K</span>
      </div>
      <nav className="mt-3 flex flex-col gap-1 text-[13px]">
        <Link to="/cases" className="flex items-center justify-between rounded-md px-3 py-2 opacity-75 hover:bg-sidebar-accent" activeProps={{ className: 'bg-sidebar-accent font-medium !opacity-100' }}>
          <span>{uk.nav.cases}</span>
          {casesCount !== undefined && <span className="text-xs opacity-50">{casesCount}</span>}
        </Link>
        {soon.map((label) => (
          <span key={label} title={uk.nav.soon} aria-disabled className="flex cursor-default items-center justify-between rounded-md px-3 py-2 opacity-40">
            {label}
          </span>
        ))}
      </nav>
    </aside>
  );
}

export function TopBar({ crumbs, end }: { crumbs: { label: string; to?: string }[]; end?: React.ReactNode }) {
  return (
    <div className="flex items-center gap-2 border-b bg-card px-4 py-3 text-[13px] lg:px-8">
      <BrandMark className="lg:hidden" />
      <nav aria-label="Breadcrumbs" className="flex min-w-0 flex-1 items-center gap-2">
        {crumbs.map((c, i) => {
          const last = i === crumbs.length - 1;
          // Mobile keeps the first and the last crumb only.
          const middle = i > 0 && !last;
          return (
            <span key={c.label} className={cn('flex min-w-0 items-center gap-2', middle && 'max-lg:hidden')}>
              {i > 0 && <span className="text-muted-foreground">/</span>}
              {c.to && !last ? (
                <Link to={c.to} className="truncate text-muted-foreground hover:text-foreground">
                  {c.label}
                </Link>
              ) : (
                <span className={cn('truncate', last ? 'font-medium' : 'text-muted-foreground')}>{c.label}</span>
              )}
            </span>
          );
        })}
      </nav>
      {end}
    </div>
  );
}
