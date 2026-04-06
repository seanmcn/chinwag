// Ports of the Go template helpers in internal/render/render.go.

export const comma = (n: number): string =>
  (n ?? 0).toLocaleString('en-US');

export const commaF = (n: number): string =>
  (n ?? 0).toLocaleString('en-US', { maximumFractionDigits: 1 });

export const fmtSec = (s: number): string => {
  if (!s || s <= 0) return '—';
  if (s < 60) return `${s}s`;
  const m = Math.round(s / 60);
  if (m < 60) return `${m}m`;
  const h = Math.floor(m / 60);
  const mm = m % 60;
  return mm ? `${h}h ${mm}m` : `${h}h`;
};

export const fmtHour = (h: number): string =>
  `${String(h ?? 0).padStart(2, '0')}:00`;

export const initial = (s: string): string =>
  (s || '?').trim().charAt(0).toUpperCase();

// Comparison helpers — return CSS class names for win/lose chips.
export const cmpA = (a: number, b: number): string => (a > b ? 'chip win' : a < b ? 'chip lose' : 'chip');
export const cmpB = (a: number, b: number): string => (b > a ? 'chip win' : b < a ? 'chip lose' : 'chip');
// Smaller-is-better variants (e.g. response times)
export const cmpAS = (a: number, b: number): string => (a > 0 && (b === 0 || a < b) ? 'chip win' : a > b ? 'chip lose' : 'chip');
export const cmpBS = (a: number, b: number): string => (b > 0 && (a === 0 || b < a) ? 'chip win' : b > a ? 'chip lose' : 'chip');
