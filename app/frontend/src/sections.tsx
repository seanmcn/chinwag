import type { analyse } from '../wailsjs/go/models';
import { comma, fmtSec, initial, cmpA, cmpB, cmpAS, cmpBS } from './format';
import { GrowthChart, SentimentChart, ReplySpeedChart, SankeyChart, Heatmap, DailyActivity } from './charts';

type S = analyse.Stats;

function Avatar({ name, who }: { name: string; who: 'me' | 'them' }) {
  return <span className={`av av-${who}`}>{initial(name)}</span>;
}

function Card({ title, sub, children, wide }: { title: string; sub?: string; children: any; wide?: boolean }) {
  return (
    <div className={`card${wide ? ' wide' : ''}`}>
      <h2>{title}</h2>
      {sub && <div className="sub">{sub}</div>}
      {children}
    </div>
  );
}

export function Topbar({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const start = new Date(stats.Period.Start).toLocaleDateString();
  const end = new Date(stats.Period.End).toLocaleDateString();
  return (
    <div className="topbar">
      <div className="pill"><div className="pill-num">{comma(stats.Messages)}</div><div className="pill-lbl">messages</div></div>
      <div className="pill"><div className="pill-num">{comma(stats.Conversations)}</div><div className="pill-lbl">conversations</div></div>
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
      <div className="pill"><div className="pill-num">{stats.Rating}/100</div><div className="pill-lbl">{stats.RatingLabel || 'rating'}</div></div>
      <div className="pill"><div className="pill-num">{stats.LongestStreak}d</div><div className="pill-lbl">longest streak</div></div>
    </div>
  );
}

export function Overview({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const meU = stats.PerUser[me]; const themU = stats.PerUser[them];
  if (!meU || !themU) return null;
  return (
    <div className="grid">
      <Card title="Volume">
        <table className="kv">
          <thead className="kv-head"><tr><th></th><th><Avatar name={me} who="me" />{me}</th><th><Avatar name={them} who="them" />{them}</th></tr></thead>
          <tbody>
            <tr><td>Messages</td><td><span className={cmpA(meU.Messages, themU.Messages)}>{comma(meU.Messages)}</span></td><td><span className={cmpB(meU.Messages, themU.Messages)}>{comma(themU.Messages)}</span></td></tr>
            <tr><td>Words</td><td><span className={cmpA(meU.Words, themU.Words)}>{comma(meU.Words)}</span></td><td><span className={cmpB(meU.Words, themU.Words)}>{comma(themU.Words)}</span></td></tr>
            <tr><td>Characters</td><td><span className={cmpA(meU.Characters, themU.Characters)}>{comma(meU.Characters)}</span></td><td><span className={cmpB(meU.Characters, themU.Characters)}>{comma(themU.Characters)}</span></td></tr>
            <tr><td>Questions</td><td><span className={cmpA(meU.Questions, themU.Questions)}>{comma(meU.Questions)}</span></td><td><span className={cmpB(meU.Questions, themU.Questions)}>{comma(themU.Questions)}</span></td></tr>
            <tr><td>Laughs</td><td><span className={cmpA(meU.Laughs, themU.Laughs)}>{comma(meU.Laughs)}</span></td><td><span className={cmpB(meU.Laughs, themU.Laughs)}>{comma(themU.Laughs)}</span></td></tr>
          </tbody>
        </table>
      </Card>
      <Card title="Response Times" sub="median / p90 — lower is faster">
        <table className="kv">
          <thead className="kv-head"><tr><th></th><th>{me}</th><th>{them}</th></tr></thead>
          <tbody>
            <tr><td>Median</td><td><span className={cmpAS(meU.MedianResp, themU.MedianResp)}>{fmtSec(meU.MedianResp)}</span></td><td><span className={cmpBS(meU.MedianResp, themU.MedianResp)}>{fmtSec(themU.MedianResp)}</span></td></tr>
            <tr><td>p90</td><td><span className={cmpAS(meU.P90Response, themU.P90Response)}>{fmtSec(meU.P90Response)}</span></td><td><span className={cmpBS(meU.P90Response, themU.P90Response)}>{fmtSec(themU.P90Response)}</span></td></tr>
            <tr><td>Average</td><td><span className={cmpAS(meU.AvgResponse, themU.AvgResponse)}>{fmtSec(meU.AvgResponse)}</span></td><td><span className={cmpBS(meU.AvgResponse, themU.AvgResponse)}>{fmtSec(themU.AvgResponse)}</span></td></tr>
          </tbody>
        </table>
      </Card>
      <Card title="Media">
        <table className="kv">
          <thead className="kv-head"><tr><th></th><th>{me}</th><th>{them}</th></tr></thead>
          <tbody>
            <tr><td>Images</td><td><span className={cmpA(meU.Images, themU.Images)}>{comma(meU.Images)}</span></td><td><span className={cmpB(meU.Images, themU.Images)}>{comma(themU.Images)}</span></td></tr>
            <tr><td>Videos</td><td><span className={cmpA(meU.Videos, themU.Videos)}>{comma(meU.Videos)}</span></td><td><span className={cmpB(meU.Videos, themU.Videos)}>{comma(themU.Videos)}</span></td></tr>
            <tr><td>Audio</td><td><span className={cmpA(meU.Audios, themU.Audios)}>{comma(meU.Audios)}</span></td><td><span className={cmpB(meU.Audios, themU.Audios)}>{comma(themU.Audios)}</span></td></tr>
            <tr><td>Stickers / GIFs</td><td><span className={cmpA(meU.Stickers + meU.GIFs, themU.Stickers + themU.GIFs)}>{comma(meU.Stickers + meU.GIFs)}</span></td><td><span className={cmpB(meU.Stickers + meU.GIFs, themU.Stickers + themU.GIFs)}>{comma(themU.Stickers + themU.GIFs)}</span></td></tr>
            <tr><td>Links</td><td><span className={cmpA(meU.Links, themU.Links)}>{comma(meU.Links)}</span></td><td><span className={cmpB(meU.Links, themU.Links)}>{comma(themU.Links)}</span></td></tr>
          </tbody>
        </table>
      </Card>
    </div>
  );
}

export function Activity({ stats }: { stats: S }) {
  return (
    <>
      <Card title="Growth over time" wide><GrowthChart stats={stats} /></Card>
      <Card title="Messaging times" wide><Heatmap stats={stats} /></Card>
      <Card title="Daily activity" wide><DailyActivity stats={stats} /></Card>
    </>
  );
}

export function Conversations({ stats }: { stats: S }) {
  return (
    <>
      <Card title="Conversation flow" wide><SankeyChart stats={stats} /></Card>
      <Card title="Reply speed by hour" wide><ReplySpeedChart stats={stats} /></Card>
    </>
  );
}

export function Sentiment({ stats }: { stats: S }) {
  return <Card title="Sentiment over time" wide><SentimentChart stats={stats} /></Card>;
}

export function Topics({ stats }: { stats: S }) {
  const [me, them] = stats.Participants;
  const meU = stats.PerUser[me]; const themU = stats.PerUser[them];
  return (
    <>
      <Card title="Top terms" wide>
        <div className="topic-block">
          <div className="topic-head">Overall</div>
          <div className="tag-row">{(stats.TopTerms ?? []).map((t, i) => <span key={i} className="tag">{t}</span>)}</div>
        </div>
        {meU && (
          <div className="topic-block">
            <div className="topic-head">{me}</div>
            <div className="tag-row">{(meU.TopTerms ?? []).map((t, i) => <span key={i} className="tag tag-me">{t}</span>)}</div>
          </div>
        )}
        {themU && (
          <div className="topic-block">
            <div className="topic-head">{them}</div>
            <div className="tag-row">{(themU.TopTerms ?? []).map((t, i) => <span key={i} className="tag tag-them">{t}</span>)}</div>
          </div>
        )}
      </Card>
      <Card title="Top shared domains" wide>
        <div className="tag-row">
          {(stats.TopDomains ?? []).map((d, i) => <span key={i} className="tag">{d.domain}<b>{d.count}</b></span>)}
          {!(stats.TopDomains?.length) && <span className="tag-empty">no links shared</span>}
        </div>
      </Card>
    </>
  );
}

export function Insights({ stats }: { stats: S }) {
  return (
    <Card title="Insights" wide>
      <ul className="insights">
        {(stats.Insights ?? []).map((line, i) => (
          <li key={i}><span className="dot" /> {line}</li>
        ))}
      </ul>
    </Card>
  );
}
