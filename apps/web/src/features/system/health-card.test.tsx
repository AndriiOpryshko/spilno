import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { HealthView } from './health-card';

describe('HealthView', () => {
  it('shows status and database version when healthy', () => {
    render(<HealthView loading={false} health={{ status: 'ok', database: { status: 'up', version: '18.6' } }} />);
    expect(screen.getByText('ok')).toBeInTheDocument();
    expect(screen.getByText('up (Postgres 18.6)')).toBeInTheDocument();
  });

  it('shows degraded state without a version', () => {
    render(<HealthView loading={false} health={{ status: 'degraded', database: { status: 'down' } }} />);
    expect(screen.getByText('degraded')).toBeInTheDocument();
    expect(screen.getByText('down')).toBeInTheDocument();
  });

  it('reports an unreachable service', () => {
    render(<HealthView loading={false} health={undefined} />);
    expect(screen.getByRole('alert')).toHaveTextContent('unreachable');
  });
});
