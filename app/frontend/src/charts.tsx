import { useEffect, useRef } from 'react';
import {
  Chart,
  LineController, BarController, RadarController,
  LineElement, BarElement, PointElement, ArcElement,
  CategoryScale, LinearScale, RadialLinearScale,
  Filler, Legend, Tooltip,
} from 'chart.js';
import { SankeyController, Flow } from 'chartjs-chart-sankey';
import type { main, analyse } from '../wailsjs/go/models';
type Stats = main.StatsDTO;
type UserStats = analyse.UserStats;

Chart.register(
  LineController, BarController, RadarController,
  LineElement, BarElement, PointElement, ArcElement,
  CategoryScale, LinearScale, RadialLinearScale,
  Filler, Legend, Tooltip,
  SankeyController, Flow,
);

// NRC categories in the bit order used by internal/lexicon. Indices 8/9
// (negative/positive) are intentionally omitted from the radar — they're
// VADER-redundant and live on SentimentCard instead.
export const NRC_RADAR_CATEGORIES = [
  'anger', 'anticipation', 'disgust', 'fear', 'joy', 'sadness', 'surprise', 'trust',
] as const;

const COL_ME = '#58a6ff';
const COL_THEM = '#3fb950';

const tickColor = '#9aa0c4';
const gridColor = '#ffffff08';
const legendLabel = { color: '#eef0ff', boxWidth: 10, boxHeight: 10, usePointStyle: true, pointStyle: 'rectRounded' as const };

function useChart(make: () => Chart | undefined, deps: unknown[]) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const c = make();
    return () => { c?.destroy(); };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);
  return ref;
}

export function GrowthChart({ stats }: { stats: Stats }) {
  const [me, them] = stats.Participants;
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    if (!ref.current) return;
    const c = new Chart(ref.current, {
      type: 'line',
      data: {
        labels: stats.Timeline.map(p => p.month),
        datasets: [
          { label: me, data: stats.Timeline.map(p => p.counts?.[me] ?? 0), borderColor: COL_ME, backgroundColor: COL_ME + '22', tension: 0.35, pointRadius: 0, fill: true, borderWidth: 2 },
          { label: them, data: stats.Timeline.map(p => p.counts?.[them] ?? 0), borderColor: COL_THEM, backgroundColor: COL_THEM + '22', tension: 0.35, pointRadius: 0, fill: true, borderWidth: 2 },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { labels: legendLabel } },
        scales: {
          x: { ticks: { color: tickColor, maxTicksLimit: 14 }, grid: { color: gridColor } },
          y: { ticks: { color: tickColor }, grid: { color: gridColor } },
        },
      },
    });
    return () => c.destroy();
  }, [stats]);
  return <div className="chart-box tall"><canvas ref={ref} /></div>;
}

export function EmotionRadar({ stats }: { stats: Stats }) {
  const [me, them] = stats.Participants;
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    if (!ref.current) return;
    const meU = stats.PerUser[me], themU = stats.PerUser[them];
    if (!meU || !themU) return;
    // Normalise to share-of-messages so two users with different volumes
    // are visually comparable. Skip indices 8 and 9 (negative/positive).
    const norm = (u: UserStats) => {
      const denom = u.ScoredMsgs || 1;
      return NRC_RADAR_CATEGORIES.map((_, i) => (u.Emotion?.[i] ?? 0) / denom);
    };
    const c = new Chart(ref.current, {
      type: 'radar',
      data: {
        labels: NRC_RADAR_CATEGORIES.map(c => c[0].toUpperCase() + c.slice(1)),
        datasets: [
          { label: me, data: norm(meU), borderColor: COL_ME, backgroundColor: COL_ME + '33', pointBackgroundColor: COL_ME, borderWidth: 2 },
          { label: them, data: norm(themU), borderColor: COL_THEM, backgroundColor: COL_THEM + '33', pointBackgroundColor: COL_THEM, borderWidth: 2 },
        ],
      },
      options: {
        responsive: true,
        // Don't try to auto-size: the wrapper (.chart-box.radar) is already
        // a fixed-size square and we want the canvas to fill it exactly.
        maintainAspectRatio: false,
        plugins: {
          legend: { labels: legendLabel },
          tooltip: { callbacks: { label: ctx => `${ctx.dataset.label}: ${(ctx.parsed.r * 100).toFixed(1)}% of msgs` } },
        },
        scales: {
          r: {
            angleLines: { color: gridColor },
            grid: { color: gridColor },
            pointLabels: { color: tickColor, font: { size: 13 } },
            ticks: {
              color: tickColor,
              backdropColor: 'transparent',
              showLabelBackdrop: false,
              callback: (v: any) => `${(v * 100).toFixed(0)}%`,
            },
            beginAtZero: true,
          },
        },
      },
    });
    return () => c.destroy();
  }, [stats]);
  return <div className="chart-box radar"><canvas ref={ref} /></div>;
}

export function SentimentChart({ stats }: { stats: Stats }) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    if (!ref.current || !stats.SentimentTimeline?.length) return;
    const c = new Chart(ref.current, {
      type: 'line',
      data: {
        labels: stats.SentimentTimeline.map(p => p.month),
        datasets: [
          { label: 'Net tone', data: stats.SentimentTimeline.map(p => p.net), borderColor: '#f59e0b', backgroundColor: '#f59e0b22', tension: 0.35, pointRadius: 0, fill: true, borderWidth: 2 },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { labels: legendLabel } },
        scales: {
          x: { ticks: { color: tickColor, maxTicksLimit: 14 }, grid: { color: gridColor } },
          y: { ticks: { color: tickColor }, grid: { color: gridColor } },
        },
      },
    });
    return () => c.destroy();
  }, [stats]);
  return <div className="chart-box"><canvas ref={ref} /></div>;
}

export function ReplySpeedChart({ stats }: { stats: Stats }) {
  const [me, them] = stats.Participants;
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    if (!ref.current) return;
    const meU = stats.PerUser[me], themU = stats.PerUser[them];
    if (!meU || !themU) return;
    const labels = Array.from({ length: 24 }, (_, h) => h);
    const toMin = (arr: number[]) => (arr || []).map(s => (s ? Math.round(s / 60) : null));
    const c = new Chart(ref.current, {
      type: 'bar',
      data: {
        labels,
        datasets: [
          { label: me, data: toMin(meU.ReplyByHour) as number[], backgroundColor: COL_ME + 'cc', borderRadius: 3 },
          { label: them, data: toMin(themU.ReplyByHour) as number[], backgroundColor: COL_THEM + 'cc', borderRadius: 3 },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: {
          legend: { labels: legendLabel },
          tooltip: { callbacks: { label: ctx => `${ctx.dataset.label}: ${ctx.parsed.y} min` } },
        },
        scales: {
          x: { ticks: { color: tickColor, callback: (v: any) => v + ':00' }, grid: { color: gridColor } },
          y: { ticks: { color: tickColor, callback: (v: any) => v + 'm' }, grid: { color: gridColor }, beginAtZero: true },
        },
      },
    });
    return () => c.destroy();
  }, [stats]);
  return <div className="chart-box"><canvas ref={ref} /></div>;
}

export function SankeyChart({ stats }: { stats: Stats }) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    if (!ref.current || !stats.Convos?.Sankey?.length) return;
    const flows = stats.Convos.Sankey.map(l => ({ from: l.source, to: l.target, flow: l.value }));
    const c = new Chart(ref.current, {
      type: 'sankey' as any,
      data: {
        datasets: [{
          data: flows as any,
          colorFrom: () => COL_ME,
          colorTo: () => COL_THEM,
          colorMode: 'gradient',
          color: '#eef0ff',
        } as any],
      },
      options: { plugins: { legend: { display: false } }, responsive: true, maintainAspectRatio: false } as any,
    });
    return () => c.destroy();
  }, [stats]);
  return <div className="chart-box tall"><canvas ref={ref} /></div>;
}

export function Heatmap({ stats }: { stats: Stats }) {
  const days = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'];
  let max = 1;
  for (let d = 0; d < 7; d++) for (let h = 0; h < 24; h++) if ((stats.Heatmap?.[d]?.[h] ?? 0) > max) max = stats.Heatmap[d][h];
  const cells: JSX.Element[] = [];
  cells.push(<div key="corner" />);
  for (let h = 0; h < 24; h++) {
    cells.push(<div key={`hh${h}`} className="hh">{(h === 0 || h === 6 || h === 12 || h === 18) ? h : ''}</div>);
  }
  for (let d = 0; d < 7; d++) {
    cells.push(<div key={`d${d}`} className="dd">{days[d]}</div>);
    for (let h = 0; h < 24; h++) {
      const v = stats.Heatmap[d][h];
      const bg = v > 0 ? `rgba(255,181,71,${0.12 + 0.88 * (v / max)})` : undefined;
      cells.push(<div key={`c${d}-${h}`} className="cell" style={bg ? { background: bg } : undefined} title={`${days[d]} ${h}:00 — ${v}`} />);
    }
  }
  return <div id="heatmap">{cells}</div>;
}

export function DailyActivity({ stats }: { stats: Stats }) {
  const days = stats.DailyActivity ?? [];
  if (!days.length) return null;
  const padded = [...days];
  const first = new Date(padded[0].date + 'T00:00:00');
  const pad = first.getDay();
  for (let i = 0; i < pad; i++) padded.unshift({ date: '', count: -1 });
  const nz = days.map(d => d.count).filter(c => c > 0).sort((a, b) => a - b);
  const q = (i: number) => (nz.length ? nz[Math.floor(nz.length * i)] || 1 : 1);
  const t1 = q(0.25), t2 = q(0.5), t3 = q(0.85);
  return (
    <div id="daily-grid">
      <div className="inner">
        <div className="grid-rows">
          {padded.map((d, i) => {
            if (d.count < 0) return <div key={i} className="gcell" />;
            const lvl = d.count >= t3 ? 4 : d.count >= t2 ? 3 : d.count >= t1 ? 2 : d.count > 0 ? 1 : 0;
            return <div key={i} className={`gcell${lvl ? ' l' + lvl : ''}`} title={`${d.date}: ${d.count}`} />;
          })}
        </div>
      </div>
    </div>
  );
}
