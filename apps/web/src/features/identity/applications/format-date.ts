export function formatDate(value: string): string {
  const [y, m, d] = value.slice(0, 10).split('-');
  return `${d}.${m}.${y}`;
}
