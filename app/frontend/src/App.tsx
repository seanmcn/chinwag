import { useEffect, useRef, useState } from 'react';
import './App.css';
import { OpenFileDialog, DistinctAuthors, Analyse } from '../wailsjs/go/main/App';
import type { main } from '../wailsjs/go/models';
type Stats = main.StatsDTO;
import { Topbar, Overview, Chat, Activity } from './sections';

type Tab = 'overview' | 'chat' | 'activity';

const TABS: { id: Tab; label: string; icon: string }[] = [
  { id: 'overview', label: 'Overview', icon: '🏠' },
  { id: 'chat', label: 'Chat', icon: '💬' },
  { id: 'activity', label: 'Activity', icon: '📊' },
];

function App() {
  const [path, setPath] = useState('');
  const [authors, setAuthors] = useState<string[]>([]);
  const [me, setMe] = useState('');
  const [them, setThem] = useState('');
  const [stats, setStats] = useState<Stats | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<Tab>('overview');
  const mainRef = useRef<HTMLElement>(null);
  useEffect(() => { mainRef.current?.scrollTo({ top: 0 }); }, [tab]);

  async function pickFile() {
    setError('');
    try {
      const p = await OpenFileDialog();
      if (!p) return;
      setPath(p);
      setStats(null);
      const a = await DistinctAuthors(p);
      setAuthors(a);
      setMe(a[0] ?? '');
      setThem(a[1] ?? '');
    } catch (e: any) {
      setError(String(e?.message ?? e));
    }
  }

  async function run() {
    if (!path || !me || !them) return;
    setBusy(true); setError('');
    try {
      const s = await Analyse(path, me, them, 6);
      setStats(s);
      setTab('overview');
    } catch (e: any) {
      setError(String(e?.message ?? e));
    } finally {
      setBusy(false);
    }
  }

  if (!stats) {
    return (
      <div className="landing">
        <div className="landing-card">
          <h1>WhatsApp Analyse</h1>
          <p className="landing-sub">Open a chat export (.txt or .zip) to see stats, conversation patterns and sentiment.</p>
          <button className="primary" onClick={pickFile}>Open chat export…</button>
          {path && (
            <div className="picker">
              <div className="picker-path" title={path}>{path.split(/[\\/]/).pop()}</div>
              <div className="who-pills">
                <div className="who-pill"><span className="who-label">You</span><span className="who-name">{me || '—'}</span></div>
                <button className="swap" title="Swap you / them" onClick={() => { const m = me; setMe(them); setThem(m); }}>⇄</button>
                <div className="who-pill them"><span className="who-label">Them</span><span className="who-name">{them || '—'}</span></div>
              </div>
              <button className="primary" disabled={busy || !me || !them || me === them} onClick={run}>{busy ? 'Analysing…' : 'Analyse'}</button>
            </div>
          )}
          {error && <div className="error">{error}</div>}
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-title">WhatsApp Analyse</div>
        <nav>
          {TABS.map(t => (
            <button key={t.id} className={`navbtn${tab === t.id ? ' active' : ''}`} onClick={() => setTab(t.id)}>
              <span className="navico">{t.icon}</span>{t.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-foot">
          <div className="who-pills compact">
            <div className="who-pill"><span className="who-label">You</span><span className="who-name">{me}</span></div>
            <button className="swap" title="Swap you / them" disabled={busy}
              onClick={async () => { const m = me, t = them; setMe(t); setThem(m); setBusy(true); try { setStats(await Analyse(path, t, m, 6)); } finally { setBusy(false); } }}>⇄</button>
            <div className="who-pill them"><span className="who-label">Them</span><span className="who-name">{them}</span></div>
          </div>
          <button className="ghost" onClick={() => { setStats(null); setPath(''); setAuthors([]); }}>Open another…</button>
        </div>
      </aside>
      <main className="main" ref={mainRef}>
        <Topbar stats={stats} />
        {tab === 'overview' && <Overview stats={stats} />}
        {tab === 'chat' && <Chat stats={stats} />}
        {tab === 'activity' && <Activity stats={stats} />}
      </main>
    </div>
  );
}

export default App;
