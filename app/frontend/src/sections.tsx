import type { main } from '../wailsjs/go/models';
import { comma, commaF, fmtSec, fmtHour, initial, cmpA, cmpB, cmpAS, cmpBS } from './format';
import { GrowthChart, SentimentChart, ReplySpeedChart, SankeyChart, Heatmap, DailyActivity } from './charts';

type S = main.StatsDTO;

function Avatar({ name, who, sm }: { name: string; who: 'me' | 'them'; sm?: boolean }) {
  return <span className={`av av-${who}${sm ? ' sm' : ''}`}>{initial(name)}</span>;
}

function Card({ icon, title, sub, children, wide }: { icon?: string; title: string; sub?: string; children: any; wide?: boolean }) {
  return (
    <div className={`card${wide ? ' wide' : ''}`}>
      <h2>{icon && <span className="ico">{icon}</span>}{title}</h2>
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

function InsightsCard({ stats }: { stats: S }) {
  return (
    <Card icon="🧠" title="Key insights">
      <ul className="insights">
        {(stats.Insights ?? []).map((line, i) => (
          <li key={i}><span className="dot" /> {line}</li>
        ))}
      </ul>
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
      {block(me, 'me', a.TopEmojis)}
      {block(them, 'them', b.TopEmojis)}
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
          <tr><td>Upbeat msgs</td><td><span className={cmpA(a.Positive, b.Positive)}>{comma(a.Positive)}</span></td><td><span className={cmpB(a.Positive, b.Positive)}>{comma(b.Positive)}</span></td></tr>
          <tr><td>Downbeat msgs</td><td><span className={cmpB(a.Negative, b.Negative)}>{comma(a.Negative)}</span></td><td><span className={cmpA(a.Negative, b.Negative)}>{comma(b.Negative)}</span></td></tr>
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

// (kept above: VolumeCard, LanguageCard, MediaCard, TopicsCard, DomainsCard)
// The combined Chat export is defined after the conversation cards below.

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

// ── Combined Chat tab ───────────────────────────────────────────────
// One scrollable page that covers everything about the conversation
// itself: what was said, how it flowed, and how each side replies.
export function Chat({ stats }: { stats: S }) {
  return (
    <>
      <div className="grid">
        <VolumeCard stats={stats} />
        <LanguageCard stats={stats} />
        <TopEmojisCard stats={stats} />
        <MediaCard stats={stats} />
        <ResponseTimesCard stats={stats} />
        <ConvoAnalysisCard stats={stats} />
      </div>
      <Card icon="⏱️" title="Reply speed by hour" wide sub="average reply latency by hour-of-day"><ReplySpeedChart stats={stats} /></Card>
      <Card icon="🔀" title="Conversation flow" wide sub="how chats start, unfold and taper off"><SankeyChart stats={stats} /></Card>
      <Card icon="💗" title="Sentiment over time" wide sub="net upbeat-vs-downbeat tone, monthly"><SentimentChart stats={stats} /></Card>
      <TopicsCard stats={stats} />
      <DomainsCard stats={stats} />
    </>
  );
}
