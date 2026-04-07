import { useEffect, useRef, useState } from 'react';
import { Tooltip } from 'react-tooltip';
import 'react-tooltip/dist/react-tooltip.css';
import './App.css';
import { OpenFileDialog, DistinctAuthors, Analyse } from '../wailsjs/go/main/App';
import type { main } from '../wailsjs/go/models';
type Stats = main.StatsDTO;
import { Topbar, Overview, Conversation, Tone, Activity, Export } from './sections';
import { initial } from './format';

type Tab = 'overview' | 'conversation' | 'tone' | 'activity' | 'export';

const TABS: { id: Tab; label: string; icon: string }[] = [
  { id: 'overview',     label: 'Overview',     icon: '🏠' },
  { id: 'conversation', label: 'Conversation', icon: '💬' },
  { id: 'tone',         label: 'Tone',         icon: '💗' },
  { id: 'activity',     label: 'Activity',     icon: '📊' },
  { id: 'export',       label: 'Export',       icon: '📤' },
];

type ChatRecord = {
  id: string;
  path: string;
  filename: string;
  authors: string[];
  me: string;
  them: string;
  stats: Stats;
};

// Pending = a file the user has picked but not yet analysed.
type Pending = {
  path: string;
  filename: string;
  authors: string[];
  me: string;
  them: string;
};

function newId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID();
  return Math.random().toString(36).slice(2);
}

function basename(p: string): string {
  return p.split(/[\\/]/).pop() || p;
}

function PickerCard({ pending, busy, error, onSwap, onAnalyse, onCancel, embedded }: {
  pending: Pending;
  busy: boolean;
  error: string;
  onSwap: () => void;
  onAnalyse: () => void;
  onCancel?: () => void;
  embedded?: boolean;
}) {
  return (
    <div className={`picker${embedded ? ' embedded' : ''}`}>
      <div className="picker-path" title={pending.path}>{pending.filename}</div>
      <div className="who-pills">
        <div className="who-pill"><span className="who-label">You</span><span className="who-name">{pending.me || '—'}</span></div>
        <button className="swap" title="Swap you / them" onClick={onSwap}>⇄</button>
        <div className="who-pill them"><span className="who-label">Them</span><span className="who-name">{pending.them || '—'}</span></div>
      </div>
      <div className="picker-actions">
        {onCancel && <button className="ghost" onClick={onCancel}>Cancel</button>}
        <button className="primary" disabled={busy || !pending.me || !pending.them || pending.me === pending.them} onClick={onAnalyse}>
          {busy ? 'Analysing…' : 'Analyse'}
        </button>
      </div>
      {error && <div className="error">{error}</div>}
    </div>
  );
}

function App() {
  const [chats, setChats] = useState<ChatRecord[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [pending, setPending] = useState<Pending | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<Tab>('overview');
  const mainRef = useRef<HTMLElement>(null);
  // Reset scroll only when changing tab — switching chats keeps the same
  // scroll position so you can flick between exports and compare the same row.
  useEffect(() => { mainRef.current?.scrollTo({ top: 0 }); }, [tab]);

  const active = chats.find(c => c.id === activeId) ?? null;

  async function pickFile() {
    setError('');
    try {
      const p = await OpenFileDialog();
      if (!p) return;
      const a = await DistinctAuthors(p);
      setPending({ path: p, filename: basename(p), authors: a, me: a[0] ?? '', them: a[1] ?? '' });
    } catch (e: any) {
      setError(String(e?.message ?? e));
    }
  }

  async function analysePending() {
    if (!pending) return;
    setBusy(true); setError('');
    try {
      const stats = await Analyse(pending.path, pending.me, pending.them, 6);
      const rec: ChatRecord = {
        id: newId(),
        path: pending.path,
        filename: pending.filename,
        authors: pending.authors,
        me: pending.me,
        them: pending.them,
        stats,
      };
      setChats(prev => [...prev, rec]);
      setActiveId(rec.id);
      setPending(null);
      setTab('overview');
    } catch (e: any) {
      setError(String(e?.message ?? e));
    } finally {
      setBusy(false);
    }
  }

  async function swapActive() {
    if (!active) return;
    setBusy(true); setError('');
    try {
      const stats = await Analyse(active.path, active.them, active.me, 6);
      setChats(prev => prev.map(c => c.id === active.id ? { ...c, me: active.them, them: active.me, stats } : c));
    } catch (e: any) {
      setError(String(e?.message ?? e));
    } finally {
      setBusy(false);
    }
  }

  function removeChat(id: string) {
    setChats(prev => {
      const next = prev.filter(c => c.id !== id);
      if (id === activeId) {
        const idx = prev.findIndex(c => c.id === id);
        const fallback = next[Math.max(0, idx - 1)] ?? null;
        setActiveId(fallback?.id ?? null);
      }
      return next;
    });
  }

  // ── Landing screen ────────────────────────────────────────────────
  if (chats.length === 0) {
    return (
      <div className="landing">
        <div className="landing-card">
          <div className="landing-hero">
            <div className="landing-eyebrow">Chat analytics, on your machine</div>
            <h1>Chinwag</h1>
            <p className="landing-tagline">See the shape of your conversations — messages, rhythms and sentiment, all from a single chat export.</p>
          </div>

          {!pending && (
            <>
              <div className="landing-cta">
                <button className="primary primary-lg" onClick={pickFile}>Open chat export…</button>
                <div className="landing-hint">Supports WhatsApp <code>.txt</code> and <code>.zip</code> exports</div>
              </div>

              <div className="landing-trust">
                <div className="trust-item">
                  <div className="trust-glyph" aria-hidden>
                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="4" y="11" width="16" height="9" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/></svg>
                  </div>
                  <div className="trust-title">Private by design</div>
                  <div className="trust-desc">Your chats never leave this device.</div>
                </div>
                <div className="trust-item">
                  <div className="trust-glyph" aria-hidden>
                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="14" rx="2"/><path d="M8 21h8M12 18v3"/></svg>
                  </div>
                  <div className="trust-title">Runs locally</div>
                  <div className="trust-desc">All parsing and analysis happen offline.</div>
                </div>
                <div className="trust-item">
                  <div className="trust-glyph" aria-hidden>
                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
                  </div>
                  <div className="trust-title">No accounts</div>
                  <div className="trust-desc">Nothing to sign up for, nothing tracked.</div>
                </div>
              </div>
            </>
          )}

          {pending && (
            <PickerCard
              pending={pending}
              busy={busy}
              error={error}
              onSwap={() => setPending({ ...pending, me: pending.them, them: pending.me })}
              onAnalyse={analysePending}
              onCancel={() => { setPending(null); setError(''); }}
            />
          )}
          {!pending && error && <div className="error">{error}</div>}
        </div>
      </div>
    );
  }

  // ── Main app ──────────────────────────────────────────────────────
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-title">Chinwag</div>

        <div className="chatlist-head">CHATS</div>
        <div className="chatlist">
          {chats.map(c => (
            <div key={c.id} className={`chatlist-row${c.id === activeId ? ' active' : ''}`} onClick={() => setActiveId(c.id)}>
              <div className="chatlist-avs">
                <span className="av av-me sm">{initial(c.me)}</span>
                <span className="av av-them sm">{initial(c.them)}</span>
              </div>
              <div className="chatlist-meta">
                <div className="chatlist-name">{c.me} & {c.them}</div>
                <div className="chatlist-file">{c.filename}</div>
              </div>
              <button
                className="chatlist-remove"
                title="Remove"
                onClick={e => { e.stopPropagation(); removeChat(c.id); }}
              >×</button>
            </div>
          ))}
          <button className="chatlist-add" onClick={pickFile}>+ Add chat</button>
        </div>

        {pending && (
          <PickerCard
            pending={pending}
            busy={busy}
            error={error}
            embedded
            onSwap={() => setPending({ ...pending, me: pending.them, them: pending.me })}
            onAnalyse={analysePending}
            onCancel={() => { setPending(null); setError(''); }}
          />
        )}

        <nav>
          {TABS.map(t => (
            <button key={t.id} className={`navbtn${tab === t.id ? ' active' : ''}`} onClick={() => setTab(t.id)}>
              <span className="navico">{t.icon}</span>{t.label}
            </button>
          ))}
        </nav>

        {active && (
          <div className="sidebar-foot">
            <div className="who-pills compact">
              <div className="who-pill"><span className="who-label">You</span><span className="who-name">{active.me}</span></div>
              <button className="swap" title="Swap you / them" disabled={busy} onClick={swapActive}>⇄</button>
              <div className="who-pill them"><span className="who-label">Them</span><span className="who-name">{active.them}</span></div>
            </div>
          </div>
        )}
      </aside>
      <main className="main" ref={mainRef}>
        {active && (
          <>
            <Topbar stats={active.stats} />
            {tab === 'overview' && <Overview stats={active.stats} />}
            {tab === 'conversation' && <Conversation stats={active.stats} />}
            {tab === 'tone' && <Tone stats={active.stats} />}
            {tab === 'activity' && <Activity stats={active.stats} />}
            {tab === 'export' && <Export stats={active.stats} />}
          </>
        )}
      </main>
      <Tooltip id="tip" className="cw-tooltip" place="top" delayShow={150} />
    </div>
  );
}

export default App;
