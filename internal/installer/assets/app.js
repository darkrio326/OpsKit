const statePrefix = '../state';
let artifactFilter = parseFilterFromURL();
let recoverCountdownTimer = null;

function htmlEscape(input) {
  return String(input)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

async function loadJson(name) {
  const res = await fetch(`${statePrefix}/${name}.json`, { cache: 'no-store' });
  if (!res.ok) throw new Error(`failed to load ${name}.json`);
  return res.json();
}

async function loadOptionalJson(path) {
  try {
    const res = await fetch(path, { cache: 'no-store' });
    if (!res.ok) return null;
    return await res.json();
  } catch (_) {
    return null;
  }
}

function card(title, bodyHtml) {
  return `<article class="card"><h3>${htmlEscape(title)}</h3>${bodyHtml}</article>`;
}

function renderOverview(overall) {
  const target = document.getElementById('overview');
  const template = (overall.activeTemplates || []).join(', ') || '-';
  const recover = overall.recoverSummary || {};
  const trend = `ok:${recover.successCount || 0} fail:${recover.failureCount || 0} warn:${recover.warnCount || 0}`;
  const lastRecover = recover.lastStatus ? `${recover.lastStatus}${recover.lastTrigger ? ` · ${recover.lastTrigger}` : ''}` : '-';
  const cooldown = cooldownLabel(recover.cooldownUntil);
  target.innerHTML = [
    card('Overall', `<div class="status ${htmlEscape(overall.overallStatus || 'UNKNOWN')}">${htmlEscape(overall.overallStatus || 'UNKNOWN')}</div>`),
    card('Open Issues', `<div>${htmlEscape(overall.openIssuesCount ?? 0)}</div>`),
    card('Template', `<div>${htmlEscape(template)}</div>`),
    card('Recover Trend', `<div>${htmlEscape(trend)}</div><div class="muted">${htmlEscape(lastRecover)}</div><div class="muted">cooldown: ${htmlEscape(cooldown)}</div>`),
    card('Last Refresh', `<div class="muted">${htmlEscape(overall.lastRefreshTime || '-')}</div>`),
  ].join('');
}

function renderRecoverAlert(overall) {
  const target = document.getElementById('recover-alert');
  const recover = overall.recoverSummary || {};
  if (!recover || (!recover.circuitOpen && !recover.lastStatus && !recover.cooldownUntil)) {
    target.innerHTML = '';
    return;
  }
  if (recover.circuitOpen && isCooldownOpen(recover.cooldownUntil)) {
    const cooldown = recover.cooldownUntil || '-';
    target.innerHTML = `<div class="alert warn"><strong>Recover circuit open.</strong> Auto-recover cooldown ${htmlEscape(cooldownLabel(cooldown))} (until ${htmlEscape(cooldown)}).</div>`;
    return;
  }
  if (recover.lastStatus === 'FAILED' || recover.lastStatus === 'WARN') {
    const trigger = recover.lastTrigger || '-';
    target.innerHTML = `<div class="alert info">Last recover status: <strong>${htmlEscape(recover.lastStatus)}</strong> (trigger: ${htmlEscape(trigger)}).</div>`;
    return;
  }
  target.innerHTML = '';
}

function renderStages(lifecycle, artifacts) {
  const stages = ['A', 'B', 'C', 'D', 'E', 'F'];
  const target = document.getElementById('stages');
  const latestRecoverCollect = latestBy(
    (x) => {
      const path = String(x.path || '').toLowerCase();
      const id = String(x.id || '').toLowerCase();
      return path.includes('collect-e-') || id === 'collect-e';
    },
    artifacts?.bundles || [],
  );
  const latestRecoverReport = latestBy(
    (x) => {
      const path = String(x.path || '').toLowerCase();
      const id = String(x.id || '').toLowerCase();
      return path.includes('recover-') || id === 'recover';
    },
    artifacts?.reports || [],
  );
  target.innerHTML = stages.map((id) => {
    const s = (lifecycle.stages || []).find((x) => x.stageId === id) || {
      stageId: id,
      name: 'N/A',
      status: 'NOT_STARTED',
      lastRunTime: '-',
      issues: [],
    };
    const showRecoverLinks = id === 'E' && (s.status === 'FAILED' || s.status === 'WARN');
    const recoverTrigger = id === 'E' ? stageMetricValue(s, 'recover_trigger') : '';
    const recoverLinks = [];
    if (showRecoverLinks && latestRecoverCollect) {
      recoverLinks.push(`<a href="${htmlEscape(linkFor(latestRecoverCollect.path))}" target="_blank" rel="noopener">latest collect bundle</a>`);
    }
    if (showRecoverLinks && latestRecoverReport) {
      recoverLinks.push(`<a href="${htmlEscape(linkFor(latestRecoverReport.path))}" target="_blank" rel="noopener">latest recover report</a>`);
    }
    const recoverLinksHtml = recoverLinks.length ? `<div class="stage-action">${recoverLinks.join('<span class="sep">·</span>')}</div>` : '';
    const triggerHtml = recoverTrigger ? `<div class="muted">trigger: ${htmlEscape(recoverTrigger)}</div>` : '';
    return card(`${s.stageId} ${s.name}`, `
      <div class="status ${htmlEscape(s.status)}">${htmlEscape(s.status)}</div>
      <div class="muted">${htmlEscape(s.lastRunTime || '-')}</div>
      <div class="muted">issues: ${(s.issues || []).length}</div>
      ${triggerHtml}
      ${recoverLinksHtml}
    `);
  }).join('');
}

function stageMetricValue(stage, label) {
  const metric = (stage.metrics || []).find((m) => m.label === label);
  return metric ? metric.value : '';
}

function renderServices(services) {
  const target = document.getElementById('services');
  const rows = (services.services || []).map((svc) => {
    const checks = (svc.checks || []).map((c) => `${c.checkId}:${c.result}`).join(', ');
    return `<li><strong>${htmlEscape(svc.serviceId || '-')}</strong> (${htmlEscape(svc.health || 'unknown')}) - ${htmlEscape(checks || 'no checks')}</li>`;
  });
  target.innerHTML = rows.length ? `<ul>${rows.join('')}</ul>` : '<div class="muted">no services yet</div>';
}

function linkFor(path) {
  if (!path) return '#';
  if (path.startsWith('/') || path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  return `../${path}`;
}

function renderArtifacts(artifacts, summary) {
  const target = document.getElementById('artifacts');
  const filtered = filterArtifacts(artifacts, artifactFilter);
  const reports = [...filtered.reports].reverse();
  const bundles = [...filtered.bundles].reverse();
  const reportRows = reports.map((r) => `<li>report <a href="${htmlEscape(linkFor(r.path))}" target="_blank" rel="noopener">${htmlEscape(r.id || r.path)}</a></li>`);
  const bundleRows = bundles.map((b) => `<li>bundle <a href="${htmlEscape(linkFor(b.path))}" target="_blank" rel="noopener">${htmlEscape(b.id || b.path)}</a></li>`);
  const sections = [];
  if (summary) {
    const status = htmlEscape(summary.overall || 'unknown');
    const failed = htmlEscape(summary.failedSteps ?? 0);
    const total = htmlEscape(summary.totalSteps ?? 0);
    sections.push(`<h3>Session</h3><ul><li>generic e2e <a href="../summary.json" target="_blank" rel="noopener">summary.json</a> (${status}, failed ${failed}/${total})</li></ul>`);
  }
  if (reportRows.length) {
    sections.push(`<h3>Reports</h3><ul>${reportRows.join('')}</ul>`);
  }
  if (bundleRows.length) {
    sections.push(`<h3>Bundles</h3><ul>${bundleRows.join('')}</ul>`);
  }
  target.innerHTML = sections.length ? sections.join('') : '<div class="muted">no artifacts yet</div>';
}

function artifactKind(item) {
  const path = String(item?.path || '').toLowerCase();
  const id = String(item?.id || '').toLowerCase();
  if (path.includes('collect-e-') || id === 'collect-e') return 'recover';
  if (path.includes('acceptance-') || id.includes('accept')) return 'acceptance';
  if (path.includes('handover-') || id.includes('handover')) return 'handover';
  return 'other';
}

function filterArtifacts(artifacts, filter) {
  const reports = artifacts.reports || [];
  const bundles = artifacts.bundles || [];
  if (filter === 'all') return { reports, bundles };
  return {
    reports: reports.filter((x) => artifactKind(x) === filter),
    bundles: bundles.filter((x) => artifactKind(x) === filter),
  };
}

function renderArtifactFilters(artifacts) {
  const target = document.getElementById('artifact-filters');
  const defs = [
    { id: 'all', label: 'All' },
    { id: 'recover', label: 'Recover Collect' },
    { id: 'acceptance', label: 'Acceptance' },
    { id: 'handover', label: 'Handover' },
  ];
  const count = (id) => {
    if (id === 'all') return (artifacts.reports || []).length + (artifacts.bundles || []).length;
    const filtered = filterArtifacts(artifacts, id);
    return filtered.reports.length + filtered.bundles.length;
  };
  target.innerHTML = defs.map((d) => {
    const active = d.id === artifactFilter ? 'active' : '';
    return `<button class="filter-btn ${active}" data-filter="${htmlEscape(d.id)}">${htmlEscape(d.label)} (${count(d.id)})</button>`;
  }).join('');
  target.querySelectorAll('.filter-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      artifactFilter = btn.dataset.filter || 'all';
      syncFilterToURL(artifactFilter);
      renderArtifactFilters(artifacts);
      renderArtifacts(artifacts);
    });
  });
}

function allowedFilter(id) {
  return ['all', 'recover', 'acceptance', 'handover'].includes(id);
}

function parseFilterFromURL() {
  const params = new URLSearchParams(window.location.search || '');
  const raw = (params.get('artifactFilter') || 'all').trim().toLowerCase();
  return allowedFilter(raw) ? raw : 'all';
}

function syncFilterToURL(filter) {
  const url = new URL(window.location.href);
  if (!allowedFilter(filter) || filter === 'all') {
    url.searchParams.delete('artifactFilter');
  } else {
    url.searchParams.set('artifactFilter', filter);
  }
  window.history.replaceState(null, '', `${url.pathname}${url.search}${url.hash}`);
}

function latestBy(predicate, items) {
  for (let i = items.length - 1; i >= 0; i -= 1) {
    if (predicate(items[i])) return items[i];
  }
  return null;
}

function cooldownLabel(until) {
  if (!until) return '-';
  const untilMs = Date.parse(until);
  if (Number.isNaN(untilMs)) return until;
  const diff = untilMs - Date.now();
  if (diff <= 0) return 'expired';
  const totalSeconds = Math.floor(diff / 1000);
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  if (h > 0) return `${h}h ${m}m ${s}s`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}

function isCooldownOpen(until) {
  if (!until) return false;
  const untilMs = Date.parse(until);
  if (Number.isNaN(untilMs)) return false;
  return untilMs > Date.now();
}

function startRecoverCountdown(overall) {
  if (recoverCountdownTimer) {
    clearInterval(recoverCountdownTimer);
    recoverCountdownTimer = null;
  }
  const recover = overall.recoverSummary || {};
  if (!recover.circuitOpen || !isCooldownOpen(recover.cooldownUntil)) {
    return;
  }
  recoverCountdownTimer = setInterval(() => {
    renderRecoverAlert(overall);
    renderOverview(overall);
    if (!isCooldownOpen(recover.cooldownUntil)) {
      clearInterval(recoverCountdownTimer);
      recoverCountdownTimer = null;
    }
  }, 1000);
}

function artifactLink(label, artifact) {
  if (!artifact) return `<span class="muted">${htmlEscape(label)}: -</span>`;
  const href = htmlEscape(linkFor(artifact.path));
  const text = htmlEscape(artifact.id || artifact.path);
  return `<a class="artifact-link" href="${href}" target="_blank" rel="noopener">${htmlEscape(label)}: ${text}</a>`;
}

function renderArtifactHighlights(artifacts, summary) {
  const target = document.getElementById('artifact-highlights');
  const reports = artifacts.reports || [];
  const bundles = artifacts.bundles || [];
  const latestAccept = latestBy((x) => (x.path || '').includes('acceptance-') || String(x.id || '').includes('accept'), bundles);
  const latestCollect = latestBy((x) => (x.path || '').includes('collect-') || String(x.id || '').includes('collect'), bundles);
  const latestHandover = latestBy((x) => (x.path || '').includes('handover-') || String(x.id || '').includes('handover'), bundles);
  const summaryCard = summary
    ? card('Generic E2E', `<a class="artifact-link" href="../summary.json" target="_blank" rel="noopener">summary.json</a><div class="muted">overall: ${htmlEscape(summary.overall || 'unknown')} · failed ${htmlEscape(summary.failedSteps ?? 0)}/${htmlEscape(summary.totalSteps ?? 0)}</div>`)
    : card('Generic E2E', '<span class="muted">summary.json not found</span>');
  target.innerHTML = [
    card('Counts', `<div>reports: ${reports.length}</div><div>bundles: ${bundles.length}</div>`),
    card('Latest Acceptance', artifactLink('bundle', latestAccept)),
    card('Latest Recover Collect', artifactLink('bundle', latestCollect)),
    card('Latest Handover', artifactLink('bundle', latestHandover)),
    summaryCard,
  ].join('');
}

async function boot() {
  const headline = document.getElementById('headline');
  try {
    const [overall, lifecycle, services, artifacts, summary] = await Promise.all([
      loadJson('overall'),
      loadJson('lifecycle'),
      loadJson('services'),
      loadJson('artifacts'),
      loadOptionalJson('../summary.json'),
    ]);

    headline.textContent = `Overall ${overall.overallStatus || 'UNKNOWN'} · refreshed ${overall.lastRefreshTime || '-'}`;
    renderRecoverAlert(overall);
    renderOverview(overall);
    renderStages(lifecycle, artifacts);
    renderServices(services);
    renderArtifactHighlights(artifacts, summary);
    renderArtifactFilters(artifacts);
    renderArtifacts(artifacts, summary);
    startRecoverCountdown(overall);
  } catch (err) {
    headline.textContent = `Failed to load state JSON: ${err.message}`;
  }
}

boot();
