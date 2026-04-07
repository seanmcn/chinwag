import html2canvas from 'html2canvas';
import { createRoot, type Root } from 'react-dom/client';
import { type ReactElement } from 'react';
import { SaveJpeg } from '../wailsjs/go/main/App';

const APP_BG = '#0e1116';
const RENDER_WIDTH = 1280;

export type ExportTabKey = 'overview' | 'conversation' | 'tone' | 'activity';

export type ExportSelection = {
  enabled: boolean;
  alias: string;
};

export function sanitiseFilename(name: string): string {
  return name
    .replace(/[\\/\u0000-\u001f]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '')
    || 'export';
}

function nextFrame(): Promise<void> {
  return new Promise(resolve => requestAnimationFrame(() => resolve()));
}

export async function captureElementToJpeg(node: HTMLElement, alias: string): Promise<string> {
  const canvas = await html2canvas(node, {
    scale: 2,
    backgroundColor: APP_BG,
    useCORS: true,
    logging: false,
    windowWidth: RENDER_WIDTH,
  });
  const dataUrl = canvas.toDataURL('image/jpeg', 0.95);
  return await SaveJpeg(`${sanitiseFilename(alias)}.jpg`, dataUrl);
}

export async function renderAndExport(
  element: ReactElement,
  alias: string,
): Promise<string> {
  const host = document.createElement('div');
  host.style.position = 'fixed';
  host.style.left = '-10000px';
  host.style.top = '0';
  host.style.width = `${RENDER_WIDTH}px`;
  host.style.background = APP_BG;
  host.style.padding = '18px';
  host.style.color = '#e6edf3';
  document.body.appendChild(host);

  let root: Root | null = null;
  try {
    root = createRoot(host);
    root.render(element);
    // Two frames so chart.js gets a chance to draw.
    await nextFrame();
    await nextFrame();
    // chart.js (esp. the Sankey on Conversation) animates for ~1s after mount.
    await new Promise(r => setTimeout(r, 1400));
    return await captureElementToJpeg(host, alias);
  } finally {
    if (root) root.unmount();
    host.remove();
  }
}

export async function exportTabs(
  tabs: { key: ExportTabKey; alias: string; element: ReactElement }[],
  onProgress?: (key: ExportTabKey) => void,
): Promise<string[]> {
  const saved: string[] = [];
  for (const t of tabs) {
    onProgress?.(t.key);
    const path = await renderAndExport(t.element, t.alias);
    if (path) saved.push(path);
  }
  return saved;
}
