import { useState } from 'react';
import type { main } from '../wailsjs/go/models';
import { comma, commaF, fmtSec, fmtHour, initial, cmpA, cmpB, cmpAS, cmpBS } from './format';
import { GrowthChart, SentimentChart, ReplySpeedChart, SankeyChart, Heatmap, DailyActivity, EmotionRadar, NRC_RADAR_CATEGORIES } from './charts';
import { exportTabs, renderAndExport, sanitiseFilename, PickDirectory, type ExportTabKey } from './exporter';

type S = main.StatsDTO;

function Avatar({ name, who, sm }: { name: string; who: 'me' | 'them'; sm?: boolean }) {
  return <span className={`av av-${who}${sm ? ' sm' : ''}`}>{initial(name)}</span>;
}

function Card({ icon, title, sub, children, wide, help }: { icon?: string; title: string; sub?: string; children: any; wide?: boolean; help?: string }) {
  return (
    <div className={`card${wide ? ' wide' : ''}`}>
      <h2>
        {icon && <span className="ico">{icon}</span>}
        <span>{title}</span>
        {help && <span className="help-mark" data-tooltip-id="tip" data-tooltip-content={help}>?</span>}
      </h2>
      {sub && <div className="sub">{sub}</div>}
      {children}
    </div>
  );
}

function CompareHead({ me, them }: { me: string; them: string }) {
  return (
    <thead className="kv-head">
      <tr>
        <th></th>
        <th><Avatar name={me} who="me" sm /> {me}</th>
        <th><Avatar name={them} who="them" sm /> {them}</th>
      </tr>
    </thead>
  );
}

// ── Topbar ──────────────────────────────────────────────────────────
export function Topbar({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const start = new Date(stats.Period.Start).toLocaleDateString();
  const end = new Date(stats.Period.End).toLocaleDateString();
  return (
    <div className="topbar">
      <div className="pill"><div className="pill-num">{comma(stats.ChatPoints)}</div><div className="pill-lbl">chat points</div></div>
      <div className="pill"><div className="pill-num">{comma(stats.Messages)}</div><div className="pill-lbl">messages</div></div>
      <div className="pill title">
        <div className="title-row">
          <Avatar name={me} who="me" />
          <div>
            <div className="title-name">{me} & {them}</div>
            <div className="title-sub">{start} → {end}</div>
          </div>
          <Avatar name={them} who="them" />
        </div>
      </div>
      <div className="pill"><div className="pill-num">{comma(stats.Conversations)}</div><div className="pill-lbl">conversations</div></div>
      <div className="pill"><div className="pill-num">{stats.LongestStreak}d</div><div className="pill-lbl">longest streak</div></div>
    </div>
  );
}

// ── Overview ────────────────────────────────────────────────────────
function RatingRing({ stats }: { stats: S }) {
  const C = 326.7; // 2πr where r=52
  const offset = C - (C * (stats.Rating / 100));
  return (
    <Card icon="🏆" title="Chat rating" sub="balance, speed, reciprocity">
      <div className="rating-wrap">
        <svg className="ring" width="120" height="120" viewBox="0 0 120 120">
          <circle cx="60" cy="60" r="52" stroke="#ffffff15" strokeWidth="10" fill="none" />
          <circle cx="60" cy="60" r="52" stroke="#4ade80" strokeWidth="10" fill="none"
            strokeLinecap="round" strokeDasharray={C}
            strokeDashoffset={offset.toFixed(1)}
            transform="rotate(-90 60 60)" />
          <text x="60" y="68" textAnchor="middle" fontSize="34" fontWeight="700" fill="#e8ebff">{stats.Rating}</text>
        </svg>
        <div className="rating-label">{stats.RatingLabel}</div>
      </div>
    </Card>
  );
}

function BalanceCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const mePts = stats.Balance?.[me] ?? 0;
  const themPts = stats.Balance?.[them] ?? 0;
  const mePct = stats.BalancePct?.[me] ?? 50;
  const themPct = stats.BalancePct?.[them] ?? 50;
  return (
    <Card icon="⚖️" title="Balance" sub="how the relationship is shared overall">
      <div className="balance-pts">
        <div>{comma(mePts)} pts</div>
        <div>{comma(themPts)} pts</div>
      </div>
      <div className="bar">
        <div className="left" style={{ width: `${mePct}%` }}>{mePct}%</div>
        <div className="right" style={{ width: `${themPct}%` }}>{themPct}%</div>
      </div>
      <div className="balance-foot">
        <Avatar name={me} who="me" sm />
        <span className="balance-tag">{Math.abs(mePct - themPct) <= 6 ? 'Balanced' : mePct > themPct ? `${me} carries more` : `${them} carries more`}</span>
        <Avatar name={them} who="them" sm />
      </div>
    </Card>
  );
}

function StreaksCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const meU = stats.PerUser[me]; const themU = stats.PerUser[them];
  return (
    <Card icon="🌗" title="Rhythms & streaks">
      <div className="streaks">
        <div className="streak"><div className="streak-num">{stats.LongestStreak}</div><div className="streak-lbl">Longest streak (days)</div></div>
        <div className="streak"><div className="streak-num">{stats.CurrentStreak}</div><div className="streak-lbl">Current streak (days)</div></div>
      </div>
      {meU && themU && (
        <table className="kv">
          <CompareHead me={me} them={them} />
          <tbody>
            <tr><td>Peak hour</td><td><span className="chip">{fmtHour(meU.PeakHour)}</span></td><td><span className="chip">{fmtHour(themU.PeakHour)}</span></td></tr>
            <tr><td>Chronotype</td><td><span className="chip">{meU.Chronotype}</span></td><td><span className="chip">{themU.Chronotype}</span></td></tr>
          </tbody>
        </table>
      )}
    </Card>
  );
}

const INSIGHT_GROUP_ORDER = ['initiative', 'balance', 'tone', 'energy', 'rhythm'] as const;
const INSIGHT_GROUP_LABEL: Record<string, string> = {
  initiative: 'Initiative',
  balance: 'Balance',
  tone: 'Tone',
  energy: 'Energy',
  rhythm: 'Rhythm & streaks',
};

function InsightsCard({ stats }: { stats: S }) {
  const insights = stats.Insights ?? [];
  if (insights.length === 0) {
    return (
      <Card icon="🧠" title="Key insights">
        <div className="sub">Not enough signal yet.</div>
      </Card>
    );
  }
  const groups = new Map<string, typeof insights>();
  for (const ins of insights) {
    const key = ins.Category || 'other';
    if (!groups.has(key)) groups.set(key, [] as any);
    (groups.get(key) as any).push(ins);
  }
  const ordered = [
    ...INSIGHT_GROUP_ORDER.filter(k => groups.has(k)),
    ...[...groups.keys()].filter(k => !INSIGHT_GROUP_ORDER.includes(k as any)),
  ];
  return (
    <Card icon="🧠" title="Key insights">
      <div className="insight-groups">
        {ordered.map(key => (
          <div className="insight-group" key={key}>
            <div className="insight-group-head">{INSIGHT_GROUP_LABEL[key] ?? key}</div>
            <div className="insight-grid">
              {(groups.get(key) ?? []).map((ins, i) => (
                <div className={`insight-tile tone-${ins.Tone || 'info'}`} key={i}>
                  <span className="insight-icon">{ins.Icon}</span>
                  <div className="insight-body">
                    <div className="insight-title">{ins.Title}</div>
                    {ins.Detail && <div className="insight-detail">{ins.Detail}</div>}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}

export function Overview({ stats }: { stats: S }) {
  return (
    <div className="grid">
      <RatingRing stats={stats} />
      <BalanceCard stats={stats} />
      <StreaksCard stats={stats} />
      <div style={{ gridColumn: '1 / -1' }}><InsightsCard stats={stats} /></div>
    </div>
  );
}

// ── Messages ────────────────────────────────────────────────────────
function VolumeCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const row = (label: string, av: number, bv: number, fmt: (n: number) => string = comma) => (
    <tr><td>{label}</td>
      <td><span className={cmpA(av, bv)}>{fmt(av)}</span></td>
      <td><span className={cmpB(av, bv)}>{fmt(bv)}</span></td>
    </tr>
  );
  return (
    <Card icon="💬" title="Message analysis" sub="volume and richness of writing">
      <table className="kv">
        <CompareHead me={me} them={them} />
        <tbody>
          {row('Messages', a.Messages, b.Messages)}
          {row('Words', a.Words, b.Words)}
          {row('Unique words', a.UniqueWords, b.UniqueWords)}
          {row('Characters', a.Characters, b.Characters)}
          {row('Questions', a.Questions, b.Questions)}
        </tbody>
      </table>
    </Card>
  );
}

function TopEmojisCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const block = (name: string, who: 'me' | 'them', list: { Emoji: string; Count: number }[] | null) => (
    <div className="emoji-block">
      <div className="emoji-who"><Avatar name={name} who={who} sm /> {name}</div>
      <div className="emoji-grid">
        {(list ?? []).slice(0, 5).map((e, i) => (
          <div key={i} className="emoji-tile">
            <div className="emoji-glyph">{e.Emoji}</div>
            <div className="emoji-count">{comma(e.Count)}</div>
          </div>
        ))}
      </div>
    </div>
  );
  return (
    <Card icon="🥇" title="Top emojis" sub="each side's most-used reactions">
      <div className="emoji-pair">
        {block(me, 'me', a.TopEmojis)}
        {block(them, 'them', b.TopEmojis)}
      </div>
    </Card>
  );
}

function LanguageCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  return (
    <Card icon="😀" title="Language analysis" sub="emojis, tone and patterns">
      <table className="kv">
        <CompareHead me={me} them={them} />
        <tbody>
          <tr><td>Emojis</td><td><span className={cmpA(a.Emojis, b.Emojis)}>{comma(a.Emojis)}</span></td><td><span className={cmpB(a.Emojis, b.Emojis)}>{comma(b.Emojis)}</span></td></tr>
          <tr><td>Laughs</td><td><span className={cmpA(a.Laughs, b.Laughs)}>{comma(a.Laughs)}</span></td><td><span className={cmpB(a.Laughs, b.Laughs)}>{comma(b.Laughs)}</span></td></tr>
          <tr><td>Apologies</td><td><span className={cmpA(a.Apologies, b.Apologies)}>{comma(a.Apologies)}</span></td><td><span className={cmpB(a.Apologies, b.Apologies)}>{comma(b.Apologies)}</span></td></tr>
          <tr><td>Encouragement</td><td><span className={cmpA(a.Encouragement, b.Encouragement)}>{comma(a.Encouragement)}</span></td><td><span className={cmpB(a.Encouragement, b.Encouragement)}>{comma(b.Encouragement)}</span></td></tr>
        </tbody>
      </table>
    </Card>
  );
}

function MediaCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const row = (label: string, av: number, bv: number) => (
    <tr><td>{label}</td>
      <td><span className={cmpA(av, bv)}>{comma(av)}</span></td>
      <td><span className={cmpB(av, bv)}>{comma(bv)}</span></td>
    </tr>
  );
  return (
    <Card icon="🎬" title="Media stats" sub="files, links and reactions shared">
      <table className="kv">
        <CompareHead me={me} them={them} />
        <tbody>
          {row('Images', a.Images, b.Images)}
          {row('Videos', a.Videos, b.Videos)}
          {row('Audios', a.Audios, b.Audios)}
          {row('GIFs', a.GIFs, b.GIFs)}
          {row('Stickers', a.Stickers, b.Stickers)}
          {row('Links', a.Links, b.Links)}
        </tbody>
      </table>
    </Card>
  );
}

function TopicsCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const meU = stats.PerUser[me]; const themU = stats.PerUser[them];
  return (
    <Card icon="🏷️" title="Topics & vocabulary" sub="distinctive words and what you both talk about">
      <div className="topic-block">
        <div className="topic-head">DISTINCTIVE TO {me.toUpperCase()}</div>
        <div className="tag-row">{(meU?.TopTerms ?? []).length
          ? meU!.TopTerms.map((t, i) => <span key={i} className="tag tag-me">{t}</span>)
          : <span className="tag-empty">Not enough data</span>}</div>
      </div>
      <div className="topic-block">
        <div className="topic-head">DISTINCTIVE TO {them.toUpperCase()}</div>
        <div className="tag-row">{(themU?.TopTerms ?? []).length
          ? themU!.TopTerms.map((t, i) => <span key={i} className="tag tag-them">{t}</span>)
          : <span className="tag-empty">Not enough data</span>}</div>
      </div>
      <div className="topic-block">
        <div className="topic-head">SHARED VOCABULARY</div>
        <div className="tag-row">{(stats.TopTerms ?? []).length
          ? stats.TopTerms.map((t, i) => <span key={i} className="tag">{t}</span>)
          : <span className="tag-empty">Not enough data</span>}</div>
      </div>
    </Card>
  );
}

function DomainsCard({ stats }: { stats: S }) {
  return (
    <Card icon="🔗" title="Top link domains" sub="which sites you share most often">
      <div className="tag-row">
        {(stats.TopDomains ?? []).length
          ? stats.TopDomains.map((d, i) => <span key={i} className="tag">{d.domain} <b>{d.count}</b></span>)
          : <span className="tag-empty">No links shared</span>}
      </div>
    </Card>
  );
}

// ── Tone tab cards ──────────────────────────────────────────────────
// All four read the new VADER + NRC fields populated by the analyser:
// CompoundAvg, Emotion[10], IntensitySum[10], IntensityPeak, VAD[3].

function compoundLabel(c: number): { text: string; cls: string } {
  if (c >= 0.2) return { text: 'warm', cls: 'tone-warm' };
  if (c >= 0.05) return { text: 'positive', cls: 'tone-pos' };
  if (c <= -0.2) return { text: 'guarded', cls: 'tone-cold' };
  if (c <= -0.05) return { text: 'negative', cls: 'tone-neg' };
  return { text: 'neutral', cls: 'tone-neutral' };
}

function SentimentCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const SENT_HELP = 'Each message is scored by VADER (Valence Aware Dictionary for sEntiment Reasoning). Compound is the per-message score from −1 (very negative) to +1 (very positive), averaged across the user\'s messages. Positive / neutral / negative split uses VADER\'s standard ±0.05 cutoffs.';
  const block = (name: string, who: 'me' | 'them', u: typeof a) => {
    const total = u.ScoredMsgs || 1;
    const pos = u.Positive;
    const neg = u.Negative;
    const neu = Math.max(0, total - pos - neg);
    const posPct = (pos / total) * 100;
    const neuPct = (neu / total) * 100;
    const negPct = (neg / total) * 100;
    const lbl = compoundLabel(u.CompoundAvg);
    return (
      <div className="tone-block">
        <div className="tone-head">
          <Avatar name={name} who={who} sm /> <span className="tone-name">{name}</span>
          <span className={`tone-tag ${lbl.cls}`} data-tooltip-id="tip" data-tooltip-content={`Compound score ${u.CompoundAvg.toFixed(3)}`}>{lbl.text}</span>
        </div>
        <div className="tone-num" data-tooltip-id="tip" data-tooltip-content={`Average VADER compound score across ${comma(total)} messages`}>{u.CompoundAvg >= 0 ? '+' : ''}{u.CompoundAvg.toFixed(2)}</div>
        <div className="bar-stack">
          <div className="seg pos" style={{ width: `${posPct}%` }} data-tooltip-id="tip" data-tooltip-content={`${posPct.toFixed(1)}% positive — ${comma(pos)} of ${comma(total)} messages`}>{posPct >= 8 ? `${posPct.toFixed(0)}%` : ''}</div>
          <div className="seg neu" style={{ width: `${neuPct}%` }} data-tooltip-id="tip" data-tooltip-content={`${neuPct.toFixed(1)}% neutral — ${comma(neu)} of ${comma(total)} messages`}>{neuPct >= 8 ? `${neuPct.toFixed(0)}%` : ''}</div>
          <div className="seg neg" style={{ width: `${negPct}%` }} data-tooltip-id="tip" data-tooltip-content={`${negPct.toFixed(1)}% negative — ${comma(neg)} of ${comma(total)} messages`}>{negPct >= 8 ? `${negPct.toFixed(0)}%` : ''}</div>
        </div>
        <div className="tone-foot">
          <span data-tooltip-id="tip" data-tooltip-content={`${posPct.toFixed(1)}% of ${name}'s messages`}><span className="dot pos" /> {comma(pos)} positive</span>
          <span data-tooltip-id="tip" data-tooltip-content={`${neuPct.toFixed(1)}% of ${name}'s messages`}><span className="dot neu" /> {comma(neu)} neutral</span>
          <span data-tooltip-id="tip" data-tooltip-content={`${negPct.toFixed(1)}% of ${name}'s messages`}><span className="dot neg" /> {comma(neg)} negative</span>
        </div>
      </div>
    );
  };
  return (
    <Card icon="💗" title="Sentiment" help={SENT_HELP} sub="overall warmth of each side's messages">
      <div className="tone-grid">
        {block(me, 'me', a)}
        {block(them, 'them', b)}
      </div>
    </Card>
  );
}

function VADCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const VAD_HELP = 'VAD = Valence, Arousal, Dominance — three psycholinguistic dimensions of word meaning. Scores come from the NRC VAD Lexicon (≈55,000 English words rated on each dimension). Each user\'s value is the average of every word in their messages, with the bar zoomed to a tight window so small differences are visible.';
  if ((a.VADMessages ?? 0) < 20 || (b.VADMessages ?? 0) < 20) {
    return (
      <Card icon="🧭" title="VAD profile" help={VAD_HELP} sub="warmth, energy and control in each side's word choices">
        <div className="tag-empty">Not enough vocabulary matches</div>
      </Card>
    );
  }
  const dims: Array<{ label: string; key: 0 | 1 | 2; hi: string; lo: string; help: string }> = [
    { label: 'Warmth',  key: 0, hi: 'warmer',          lo: 'cooler',           help: 'Valence — how positive vs negative the words feel. Words like "joyful", "love" score high; "miserable", "hate" score low.' },
    { label: 'Energy',  key: 1, hi: 'higher energy',   lo: 'calmer',           help: 'Arousal — how activating the words feel. Words like "ecstatic", "panic" score high; "calm", "drowsy" score low.' },
    { label: 'Control', key: 2, hi: 'more in control', lo: 'more deferential', help: 'Dominance — how powerful vs submissive the words feel. Words like "confident", "command" score high; "afraid", "weak" score low.' },
  ];
  return (
    <Card icon="🧭" title="VAD profile" help={VAD_HELP} sub="warmth, energy and control in each side's word choices">
      <div className="vad-list">
        {dims.map(d => {
          const av = a.VAD?.[d.key] ?? 0.5;
          const bv = b.VAD?.[d.key] ?? 0.5;
          const diff = av - bv;
          // Real VAD averages cluster tightly around 0.5, so a 0–1 axis
          // hides any meaningful gap. Auto-zoom to a window around the
          // two values with at least ±0.05 padding so the bar always has
          // visible breathing room and isn't all-or-nothing.
          const lo = Math.max(0, Math.min(av, bv) - 0.05);
          const hi = Math.min(1, Math.max(av, bv) + 0.05);
          const span = Math.max(0.01, hi - lo);
          const aPct = ((av - lo) / span) * 100;
          const bPct = ((bv - lo) / span) * 100;
          let note: string;
          let leader: 'me' | 'them' | 'even';
          if (Math.abs(diff) < 0.005) { note = 'about even'; leader = 'even'; }
          else if (diff > 0) { note = `${me} ${d.hi}`; leader = 'me'; }
          else { note = `${them} ${d.hi}`; leader = 'them'; }
          return (
            <div key={d.key} className="vad-row">
              <div className="vad-row-head">
                <div className="vad-label">
                  {d.label}
                  <span className="help-mark sm" data-tooltip-id="tip" data-tooltip-content={d.help}>?</span>
                </div>
                <div className="vad-note">{note}</div>
              </div>
              <div className="vad-values">
                <span className={`vad-num me${leader === 'me' ? ' lead' : ''}`} data-tooltip-id="tip" data-tooltip-content={`${me}: ${av.toFixed(3)}`}>{av.toFixed(3)}</span>
                <div className="vad-bar" data-tooltip-id="tip" data-tooltip-content={`${d.label}: ${me} ${av.toFixed(3)} vs ${them} ${bv.toFixed(3)}`}>
                  <span className="vad-axis" />
                  <span className="vad-marker me" style={{ left: `${aPct}%` }} data-tooltip-id="tip" data-tooltip-content={`${me}: ${av.toFixed(3)}`} />
                  <span className="vad-marker them" style={{ left: `${bPct}%` }} data-tooltip-id="tip" data-tooltip-content={`${them}: ${bv.toFixed(3)}`} />
                </div>
                <span className={`vad-num them${leader === 'them' ? ' lead' : ''}`} data-tooltip-id="tip" data-tooltip-content={`${them}: ${bv.toFixed(3)}`}>{bv.toFixed(3)}</span>
              </div>
              <div className="vad-range" data-tooltip-id="tip" data-tooltip-content="The bar is zoomed to a tight window so small differences are visible. These are the actual axis bounds.">{lo.toFixed(2)}<span>zoomed</span>{hi.toFixed(2)}</div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}

function IntensityCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  const INT_HELP = 'Each emotion-bearing word has a 0–1 intensity from the NRC Emotion Intensity Lexicon (e.g. "furious" = 0.98 anger, "annoyed" = 0.55 anger). Per-message intensity is summed per category, then averaged over all the user\'s messages. Peak moment is the single highest sum across any one message.';
  // Top 3 emotion intensities per user, normalised by ScoredMsgs.
  const top = (u: typeof a) => {
    const denom = u.ScoredMsgs || 1;
    const ranked = NRC_RADAR_CATEGORIES.map((cat, i) => ({
      cat,
      val: (u.IntensitySum?.[i] ?? 0) / denom,
    })).sort((x, y) => y.val - x.val).slice(0, 3);
    const max = ranked[0]?.val || 1;
    return { ranked, max };
  };
  const block = (name: string, who: 'me' | 'them', u: typeof a) => {
    const { ranked, max } = top(u);
    return (
      <div className="intensity-block">
        <div className="intensity-head"><Avatar name={name} who={who} sm /> {name}</div>
        {ranked.map(r => (
          <div key={r.cat} className="intensity-row" data-tooltip-id="tip" data-tooltip-content={`${r.cat}: average intensity ${r.val.toFixed(3)} per message`}>
            <div className="intensity-cat">{r.cat}</div>
            <div className="intensity-bar">
              <div className={`intensity-fill who-${who}`} style={{ width: `${(r.val / max) * 100}%` }} />
            </div>
            <div className="intensity-val">{r.val.toFixed(3)}</div>
          </div>
        ))}
        <div className="intensity-peak" data-tooltip-id="tip" data-tooltip-content={`Single highest summed intensity across any one of ${name}'s messages`}>
          peak moment <b>{u.IntensityPeak.toFixed(1)}</b>
        </div>
      </div>
    );
  };
  return (
    <Card icon="🔥" title="Emotional intensity" help={INT_HELP} sub="strongest feelings each side expresses">
      <div className="tone-stack">
        {block(me, 'me', a)}
        {block(them, 'them', b)}
      </div>
    </Card>
  );
}

// ── Tone tab ─────────────────────────────────────────────────────────
export function Tone({ stats }: { stats: S }) {
  return (
    <>
      <div className="grid">
        <div style={{ gridColumn: '1 / -1' }}><SentimentCard stats={stats} /></div>
      </div>
      <Card icon="💗" title="Sentiment over time" wide sub="monthly net tone per person"><SentimentChart stats={stats} /></Card>
      <div className="grid">
        <VADCard stats={stats} />
        <IntensityCard stats={stats} />
        <LanguageCard stats={stats} />
      </div>
      <div className="grid grid-2">
        <TopEmojisCard stats={stats} />
        <Card icon="🎭" title="Emotion mix" sub="share of messages carrying each Plutchik emotion (NRC EmoLex)">
          <EmotionRadar stats={stats} />
        </Card>
      </div>
    </>
  );
}

// (Conversation tab below — was the old "Chat" tab, minus tone content.)

// ── Activity ────────────────────────────────────────────────────────
export function Activity({ stats }: { stats: S }) {
  return (
    <>
      <Card icon="📈" title="Relationship growth" wide sub="messages exchanged over time"><GrowthChart stats={stats} /></Card>
      <Card icon="⏰" title="Messaging times" wide sub="when during the week and day you talk">
        <Heatmap stats={stats} />
        <div className="times-foot">{stats.TopWeekdayHr}</div>
        <div className="times-foot tiny">Characters typed: <b>{comma(stats.CharsTyped)}</b> · Time typing: <b>{stats.TimeTyping}</b></div>
      </Card>
      <Card icon="⏱️" title="Reply speed by hour" wide sub="average reply latency by hour-of-day"><ReplySpeedChart stats={stats} /></Card>
      <Card icon="📅" title="Daily chat activity" wide sub="last 500 days, GitHub-style"><DailyActivity stats={stats} /></Card>
    </>
  );
}

// ── Conversations ───────────────────────────────────────────────────
function ResponseTimesCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  return (
    <Card icon="⚡" title="Responding" sub="reply speed and rhythm">
      <table className="kv">
        <CompareHead me={me} them={them} />
        <tbody>
          <tr><td>Rapid 1st response</td><td><span className={cmpA(a.RapidFirstPct, b.RapidFirstPct)}>{a.RapidFirstPct}%</span></td><td><span className={cmpB(a.RapidFirstPct, b.RapidFirstPct)}>{b.RapidFirstPct}%</span></td></tr>
          <tr><td>Avg 1st response</td><td><span className={cmpAS(a.AvgFirstResp, b.AvgFirstResp)}>{fmtSec(a.AvgFirstResp)}</span></td><td><span className={cmpBS(a.AvgFirstResp, b.AvgFirstResp)}>{fmtSec(b.AvgFirstResp)}</span></td></tr>
          <tr><td>Avg response time</td><td><span className={cmpAS(a.AvgResponse, b.AvgResponse)}>{fmtSec(a.AvgResponse)}</span></td><td><span className={cmpBS(a.AvgResponse, b.AvgResponse)}>{fmtSec(b.AvgResponse)}</span></td></tr>
          <tr><td>Median (p50)</td><td><span className={cmpAS(a.MedianResp, b.MedianResp)}>{fmtSec(a.MedianResp)}</span></td><td><span className={cmpBS(a.MedianResp, b.MedianResp)}>{fmtSec(b.MedianResp)}</span></td></tr>
          <tr><td>Slow tail (p90)</td><td><span className={cmpAS(a.P90Response, b.P90Response)}>{fmtSec(a.P90Response)}</span></td><td><span className={cmpBS(a.P90Response, b.P90Response)}>{fmtSec(b.P90Response)}</span></td></tr>
        </tbody>
      </table>
    </Card>
  );
}

function ConvoAnalysisCard({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const a = stats.PerUser[me], b = stats.PerUser[them];
  if (!a || !b) return null;
  return (
    <Card icon="🔀" title="Conversation analysis" sub="how chats begin, unfold and close">
      <table className="kv">
        <CompareHead me={me} them={them} />
        <tbody>
          <tr><td>Convos started</td><td><span className={cmpA(a.ConvosStarted, b.ConvosStarted)}>{comma(a.ConvosStarted)}</span></td><td><span className={cmpB(a.ConvosStarted, b.ConvosStarted)}>{comma(b.ConvosStarted)}</span></td></tr>
          <tr><td>Convos closed</td><td><span className={cmpA(a.ConvosClosed, b.ConvosClosed)}>{comma(a.ConvosClosed)}</span></td><td><span className={cmpB(a.ConvosClosed, b.ConvosClosed)}>{comma(b.ConvosClosed)}</span></td></tr>
          <tr><td>Top contributor</td><td><span className={cmpA(a.TopContrib, b.TopContrib)}>{comma(a.TopContrib)}</span></td><td><span className={cmpB(a.TopContrib, b.TopContrib)}>{comma(b.TopContrib)}</span></td></tr>
          <tr><td>Avg convo points</td><td><span className="chip">{commaF(a.AvgConvoPts)}</span></td><td><span className="chip">{commaF(b.AvgConvoPts)}</span></td></tr>
          <tr><td>Reconnects</td><td><span className={cmpA(a.Reconnects, b.Reconnects)}>{comma(a.Reconnects)}</span></td><td><span className={cmpB(a.Reconnects, b.Reconnects)}>{comma(b.Reconnects)}</span></td></tr>
          <tr><td>Double messages</td><td><span className="chip">{comma(a.DoubleMsgs)}</span></td><td><span className="chip">{comma(b.DoubleMsgs)}</span></td></tr>
          <tr><td>Convos missed</td><td><span className="chip">{comma(a.ConvosMissed)}</span></td><td><span className="chip">{comma(b.ConvosMissed)}</span></td></tr>
        </tbody>
      </table>
    </Card>
  );
}

// ── Conversation tab ────────────────────────────────────────────────
// Mechanics of the conversation itself: volume, media, response times,
// flow. Tone / language affect lives on the Tone tab instead.
export function Conversation({ stats }: { stats: S }) {
  return (
    <>
      <div className="grid">
        <VolumeCard stats={stats} />
        <MediaCard stats={stats} />
        <ResponseTimesCard stats={stats} />
        <ConvoAnalysisCard stats={stats} />
        <TopicsCard stats={stats} />
        <DomainsCard stats={stats} />
      </div>
      <Card icon="🔀" title="Conversation flow" wide sub="how chats start, unfold and taper off"><SankeyChart stats={stats} /></Card>
    </>
  );
}

// ── Export tab ──────────────────────────────────────────────────────
const EXPORT_TABS: { key: ExportTabKey; label: string; icon: string }[] = [
  { key: 'overview',     label: 'Overview',     icon: '🏠' },
  { key: 'conversation', label: 'Conversation', icon: '💬' },
  { key: 'tone',         label: 'Tone',         icon: '💗' },
  { key: 'activity',     label: 'Activity',     icon: '📊' },
];

function defaultAlias(me: string, them: string, key: ExportTabKey): string {
  return sanitiseFilename(`${me}-and-${them}-${key}`);
}

type ExportMode = 'separate' | 'combined';

// Build a stats clone with the two participants renamed. All map keys
// (PerUser, Balance, BalancePct, Direction, SentimentTimeline.netByAuthor)
// are remapped from the original names to the new ones.
function renameStats(stats: S, newMe: string, newThem: string): S {
  const [oldMe, oldThem] = stats.Participants;
  if (newMe === oldMe && newThem === oldThem) return stats;
  const remap = (k: string) => (k === oldMe ? newMe : k === oldThem ? newThem : k);
  const remapMap = <V,>(m: Record<string, V> | undefined): Record<string, V> => {
    const out: Record<string, V> = {};
    for (const k of Object.keys(m ?? {})) out[remap(k)] = (m as any)[k];
    return out;
  };
  const sentiment = (stats.SentimentTimeline ?? []).map(p => ({
    ...p,
    netByAuthor: remapMap<number>(p.netByAuthor as any),
  })) as any;
  return {
    ...stats,
    Participants: [newMe, newThem],
    PerUser: remapMap(stats.PerUser as any),
    Balance: remapMap(stats.Balance as any),
    BalancePct: remapMap(stats.BalancePct as any),
    Direction: remapMap(stats.Direction as any),
    SentimentTimeline: sentiment,
  } as S;
}

const STEPS = ['Names', 'Sections', 'Output', 'Export'] as const;
type StepIdx = 0 | 1 | 2 | 3;

export function Export({ stats }: { stats: S }) {
  const [origMe, origThem] = stats.Participants;

  // Step state
  const [step, setStep] = useState<StepIdx>(0);
  const [aliasMe, setAliasMe] = useState(origMe);
  const [aliasThem, setAliasThem] = useState(origThem);
  const [enabled, setEnabled] = useState<Record<ExportTabKey, boolean>>({
    overview: true, conversation: true, tone: true, activity: true,
  });
  const [includeHeader, setIncludeHeader] = useState(true);
  const [mode, setMode] = useState<ExportMode>('combined');
  const [perTabAlias, setPerTabAlias] = useState<Record<ExportTabKey, string>>(() => ({
    overview:     defaultAlias(origMe, origThem, 'overview'),
    conversation: defaultAlias(origMe, origThem, 'conversation'),
    tone:         defaultAlias(origMe, origThem, 'tone'),
    activity:     defaultAlias(origMe, origThem, 'activity'),
  }));
  const [combinedAlias, setCombinedAlias] = useState(() => sanitiseFilename(`${origMe}-and-${origThem}-export`));

  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState<string>('');
  const [error, setError] = useState<string>('');

  const enabledTabs = EXPORT_TABS.filter(t => enabled[t.key]);
  const namesValid = aliasMe.trim() !== '' && aliasThem.trim() !== '' && aliasMe.trim() !== aliasThem.trim();
  const sectionsValid = enabledTabs.length > 0;
  const filenamesValid = mode === 'combined'
    ? combinedAlias.trim() !== ''
    : enabledTabs.every(t => perTabAlias[t.key].trim() !== '');
  const canNext: Record<StepIdx, boolean> = { 0: namesValid, 1: sectionsValid, 2: filenamesValid, 3: true };

  function bodyFor(s: S, key: ExportTabKey) {
    switch (key) {
      case 'overview':     return <Overview stats={s} />;
      case 'conversation': return <Conversation stats={s} />;
      case 'tone':         return <Tone stats={s} />;
      case 'activity':     return <Activity stats={s} />;
    }
  }

  async function runExport() {
    setBusy(true); setStatus(''); setError('');
    try {
      const renamed = renameStats(stats, aliasMe.trim(), aliasThem.trim());
      let saved: string[] = [];
      if (mode === 'separate') {
        const dir = await PickDirectory('Choose folder to save exports');
        if (!dir) {
          setStatus('Cancelled.');
          setBusy(false);
          return;
        }
        const jobs = enabledTabs.map(t => ({
          key: t.key,
          alias: perTabAlias[t.key],
          element: (
            <>
              {includeHeader && <Topbar stats={renamed} />}
              {bodyFor(renamed, t.key)}
            </>
          ),
        }));
        saved = await exportTabs(jobs, key => setStatus(`Exporting ${key}…`), dir);
      } else {
        setStatus('Rendering…');
        const combined = (
          <>
            {includeHeader && <Topbar stats={renamed} />}
            {enabledTabs.map(t => (
              <div key={t.key} style={{ marginTop: 18 }}>
                {bodyFor(renamed, t.key)}
              </div>
            ))}
          </>
        );
        const path = await renderAndExport(combined, combinedAlias);
        if (path) saved = [path];
      }
      if (saved.length === 0) {
        setStatus('Cancelled.');
      } else {
        const dir = saved[0].replace(/[\\/][^\\/]*$/, '');
        setStatus(`Saved ${saved.length} ${saved.length === 1 ? 'image' : 'images'} to ${dir}`);
      }
    } catch (e: any) {
      setError(String(e?.message ?? e));
      setStatus('');
    } finally {
      setBusy(false);
    }
  }

  // When names change, refresh default filenames (only if user hasn't customised away from previous default).
  function applyNames() {
    const me = aliasMe.trim(), them = aliasThem.trim();
    setPerTabAlias(prev => {
      const next = { ...prev };
      for (const t of EXPORT_TABS) {
        // Replace if it still matches the previous default for any prior name pair (heuristic: starts with old name).
        next[t.key] = defaultAlias(me, them, t.key);
      }
      return next;
    });
    setCombinedAlias(sanitiseFilename(`${me}-and-${them}-export`));
  }

  function goNext() {
    if (step === 0) applyNames();
    setStep(s => Math.min(3, (s + 1)) as StepIdx);
  }
  function goBack() { setStep(s => Math.max(0, (s - 1)) as StepIdx); }

  return (
    <div className="export-wizard">
      <div className="wizard-steps">
        {STEPS.map((label, i) => (
          <div key={label} className={`wizard-step${i === step ? ' active' : ''}${i < step ? ' done' : ''}`}>
            <span className="wizard-step-num">{i < step ? '✓' : i + 1}</span>
            <span className="wizard-step-label">{label}</span>
          </div>
        ))}
      </div>

      <div className="wizard-body">
        {step === 0 && (
          <div className="wizard-pane">
            <h3>Rename participants</h3>
            <p className="wizard-help">These names appear throughout the exported images. Defaults to the chat's real names.</p>
            <div className="wizard-fields">
              <label className="wizard-field">
                <span className="wizard-field-label">You</span>
                <input type="text" value={aliasMe} onChange={e => setAliasMe(e.target.value)} placeholder={origMe} />
              </label>
              <label className="wizard-field">
                <span className="wizard-field-label">Them</span>
                <input type="text" value={aliasThem} onChange={e => setAliasThem(e.target.value)} placeholder={origThem} />
              </label>
            </div>
            {!namesValid && (aliasMe || aliasThem) && (
              <div className="wizard-warn">Both names must be set and different.</div>
            )}
          </div>
        )}

        {step === 1 && (
          <div className="wizard-pane">
            <h3>What to include</h3>
            <p className="wizard-help">Pick which sections appear in the export.</p>
            <div className="wizard-tab-grid">
              {EXPORT_TABS.map(t => (
                <label key={t.key} className={`wizard-tab-card${enabled[t.key] ? ' on' : ''}`}>
                  <input
                    type="checkbox"
                    checked={enabled[t.key]}
                    onChange={e => setEnabled(prev => ({ ...prev, [t.key]: e.target.checked }))}
                  />
                  <span className="wizard-tab-ico">{t.icon}</span>
                  <span className="wizard-tab-label">{t.label}</span>
                </label>
              ))}
            </div>
            <label className="wizard-toggle">
              <input type="checkbox" checked={includeHeader} onChange={e => setIncludeHeader(e.target.checked)} />
              <span>Include summary header (chat points, messages, longest streak…)</span>
            </label>
            {!sectionsValid && <div className="wizard-warn">Pick at least one section.</div>}
          </div>
        )}

        {step === 2 && (
          <div className="wizard-pane">
            <h3>Output format</h3>
            <p className="wizard-help">A single tall image is best for sharing. Separate images give you one file per section.</p>
            <div className="wizard-mode-grid">
              <label className={`wizard-mode-card${mode === 'combined' ? ' on' : ''}`}>
                <input type="radio" name="exp-mode" checked={mode === 'combined'} onChange={() => setMode('combined')} />
                <div className="wizard-mode-title">📜 Single image</div>
                <div className="wizard-mode-desc">All selected sections stacked into one tall JPEG. Header (if on) appears once at the top.</div>
              </label>
              <label className={`wizard-mode-card${mode === 'separate' ? ' on' : ''}`}>
                <input type="radio" name="exp-mode" checked={mode === 'separate'} onChange={() => setMode('separate')} />
                <div className="wizard-mode-title">🗂 Separate images</div>
                <div className="wizard-mode-desc">One JPEG per section. Header (if on) is repeated on each.</div>
              </label>
            </div>

            <div className="wizard-filenames">
              {mode === 'combined' ? (
                <label className="wizard-field">
                  <span className="wizard-field-label">Filename</span>
                  <div className="export-alias">
                    <input type="text" value={combinedAlias} onChange={e => setCombinedAlias(e.target.value)} />
                    <span className="export-suffix">.jpg</span>
                  </div>
                </label>
              ) : (
                <div className="wizard-fields-list">
                  {enabledTabs.map(t => (
                    <label key={t.key} className="wizard-field">
                      <span className="wizard-field-label">{t.icon} {t.label}</span>
                      <div className="export-alias">
                        <input
                          type="text"
                          value={perTabAlias[t.key]}
                          onChange={e => setPerTabAlias(prev => ({ ...prev, [t.key]: e.target.value }))}
                        />
                        <span className="export-suffix">.jpg</span>
                      </div>
                    </label>
                  ))}
                </div>
              )}
            </div>
            {!filenamesValid && <div className="wizard-warn">Filenames cannot be empty.</div>}
          </div>
        )}

        {step === 3 && (
          <div className="wizard-pane">
            <h3>Ready to export</h3>
            <ul className="wizard-summary">
              <li><strong>You</strong> → {aliasMe}</li>
              <li><strong>Them</strong> → {aliasThem}</li>
              <li><strong>Sections</strong> → {enabledTabs.map(t => t.label).join(', ')}</li>
              <li><strong>Header</strong> → {includeHeader ? 'included' : 'hidden'}</li>
              <li><strong>Output</strong> → {mode === 'combined' ? `single image (${combinedAlias}.jpg)` : `${enabledTabs.length} separate images`}</li>
            </ul>
            <div className="wizard-actions-row">
              <button className="primary" disabled={busy} onClick={runExport}>
                {busy ? 'Exporting…' : 'Export'}
              </button>
              {status && <span className="export-status">{status}</span>}
            </div>
            {error && <div className="error">{error}</div>}
          </div>
        )}
      </div>

      <div className="wizard-nav">
        <button className="ghost" disabled={step === 0 || busy} onClick={goBack}>← Back</button>
        {step < 3 && (
          <button className="primary" disabled={!canNext[step] || busy} onClick={goNext}>Next →</button>
        )}
      </div>
    </div>
  );
}
