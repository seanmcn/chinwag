(function () {
  const D = window.__DATA__;
  const me = D.me, them = D.them;
  const colMe = '#3aa3ff', colThem = '#39d9c0';

  function safe(name, fn) {
    try { fn(); }
    catch (e) { console.error('[chart:' + name + ']', e); }
  }

  // ── Relationship growth (line) ──────────────────────────
  safe('growth', () => {
    new Chart(document.getElementById('growth'), {
      type: 'line',
      data: {
        labels: D.timeline.map(p => p.month),
        datasets: [
          { label: me, data: D.timeline.map(p => (p.counts && p.counts[me]) || 0), borderColor: colMe, backgroundColor: colMe + '22', tension: .35, pointRadius: 0, fill: true, borderWidth: 2 },
          { label: them, data: D.timeline.map(p => (p.counts && p.counts[them]) || 0), borderColor: colThem, backgroundColor: colThem + '22', tension: .35, pointRadius: 0, fill: true, borderWidth: 2 },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { labels: { color: '#eef0ff', boxWidth: 10, boxHeight: 10, usePointStyle: true, pointStyle: 'rectRounded' } } },
        scales: {
          x: { ticks: { color: '#9aa0c4', maxTicksLimit: 14 }, grid: { color: '#ffffff08' } },
          y: { ticks: { color: '#9aa0c4' }, grid: { color: '#ffffff08' } },
        },
      },
    });
  });

  // ── Messaging Times heatmap (CSS grid) ──────────────────
  safe('heatmap', () => {
    const el = document.getElementById('heatmap');
    if (!el) return;
    el.innerHTML = '';
    const days = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'];
    let max = 1;
    for (let d = 0; d < 7; d++) for (let h = 0; h < 24; h++) if (D.heatmap[d][h] > max) max = D.heatmap[d][h];
    // header row
    const corner = document.createElement('div'); el.appendChild(corner);
    for (let h = 0; h < 24; h++) {
      const c = document.createElement('div'); c.className = 'hh';
      c.textContent = (h === 0 || h === 6 || h === 12 || h === 18) ? h : '';
      el.appendChild(c);
    }
    for (let d = 0; d < 7; d++) {
      const lab = document.createElement('div'); lab.className = 'dd'; lab.textContent = days[d]; el.appendChild(lab);
      for (let h = 0; h < 24; h++) {
        const v = D.heatmap[d][h];
        const cell = document.createElement('div');
        cell.className = 'cell';
        if (v > 0) cell.style.background = `rgba(255,181,71,${0.12 + 0.88 * (v / max)})`;
        cell.title = `${days[d]} ${h}:00 — ${v}`;
        el.appendChild(cell);
      }
    }
  });

  // ── Conversation Flow (sankey or fallback) ──────────────
  safe('sankey', () => {
    if (!D.sankey || !D.sankey.length) return;
    let hasSankey = false;
    try { Chart.registry.getController('sankey'); hasSankey = true; } catch (_) {}
    if (!hasSankey) {
      const el = document.getElementById('sankey');
      if (el) {
        const list = document.createElement('ul');
        list.className = 'sankey-fallback';
        D.sankey.forEach(l => {
          const li = document.createElement('li');
          li.textContent = `${l.source} → ${l.target} (${l.value})`;
          list.appendChild(li);
        });
        el.replaceWith(list);
      }
      return;
    }
    const flows = D.sankey.map(l => ({ from: l.source, to: l.target, flow: l.value }));
    new Chart(document.getElementById('sankey'), {
      type: 'sankey',
      data: { datasets: [{
        data: flows,
        colorFrom: () => colMe,
        colorTo: () => colThem,
        colorMode: 'gradient',
        color: '#eef0ff',
        font: { family: 'Inter Tight', weight: '600', size: 12 },
        labels: Object.fromEntries([...new Set(flows.flatMap(f => [f.from, f.to]))].map(k => [k, k])),
      }] },
      options: { plugins: { legend: { display: false } }, responsive: true, maintainAspectRatio: false, color: '#eef0ff' },
    });
  });

  // ── Daily activity (GitHub-style grid) ──────────────────
  safe('daily', () => {
    const host = document.getElementById('daily-grid');
    if (!host || !D.daily || !D.daily.length) return;
    host.innerHTML = '';
    const days = D.daily.slice();
    // Pad start so first cell sits on a Sunday row
    const first = new Date(days[0].date + 'T00:00:00');
    const pad = first.getDay(); // 0=Sun
    for (let i = 0; i < pad; i++) days.unshift({ date: '', count: -1 });
    // Compute thresholds (quartiles of nonzero days)
    const nz = D.daily.map(d => d.count).filter(c => c > 0).sort((a, b) => a - b);
    const q = i => nz.length ? nz[Math.floor(nz.length * i)] || 1 : 1;
    const t1 = q(0.25), t2 = q(0.5), t3 = q(0.85);
    const grid = document.createElement('div');
    grid.className = 'grid-rows';
    days.forEach(d => {
      const c = document.createElement('div');
      c.className = 'gcell';
      if (d.count > 0) {
        const lvl = d.count >= t3 ? 4 : d.count >= t2 ? 3 : d.count >= t1 ? 2 : 1;
        c.classList.add('l' + lvl);
        c.title = `${d.date}: ${d.count}`;
      } else if (d.count === 0) {
        c.title = `${d.date}: 0`;
      }
      grid.appendChild(c);
    });
    const inner = document.createElement('div'); inner.className = 'inner';
    inner.appendChild(grid);
    host.appendChild(inner);
    // Months strip
    const months = document.createElement('div');
    months.className = 'months';
    const seen = new Set();
    D.daily.forEach((d, i) => {
      const m = d.date.slice(0, 7);
      if (!seen.has(m) && i % 14 === 0) {
        const span = document.createElement('span');
        const dt = new Date(d.date + 'T00:00:00');
        span.textContent = dt.toLocaleString('en', { month: 'short', year: '2-digit' });
        months.appendChild(span);
        seen.add(m);
      }
    });
    inner.appendChild(months);
  });
})();
