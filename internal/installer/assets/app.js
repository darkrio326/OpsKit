const statePrefix = '../state';
let artifactFilter = parseFilterFromURL();
let templateFilter = parseTemplateFilterFromURL();
let selectedTemplateRef = parseTemplateViewFromURL();
let selectedTemplate = null;
let recoverCountdownTimer = null;
let appState = null;

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

function wallCard(title, value, sub, extraClass = '') {
  return `<article class="wall-card ${htmlEscape(extraClass)}"><h3>${htmlEscape(title)}</h3><div class="wall-value ${htmlEscape(extraClass)}">${htmlEscape(value)}</div><div class="wall-sub">${htmlEscape(sub)}</div></article>`;
}

function currentTemplateLabel() {
  if (!selectedTemplate) {
    return 'generic-baseline';
  }
  return `${selectedTemplate.ref} (${selectedTemplate.mode}/${selectedTemplate.serviceScope})`;
}

function renderOverview(overall) {
  const target = document.getElementById('overview');
  const recover = overall.recoverSummary || {};
  const trend = `ok:${recover.successCount || 0} fail:${recover.failureCount || 0} warn:${recover.warnCount || 0}`;
  const lastRecover = recover.lastStatus ? `${recover.lastStatus}${recover.lastTrigger ? ` · ${recover.lastTrigger}` : ''}` : '-';
  const cooldown = cooldownLabel(recover.cooldownUntil);
  target.innerHTML = [
    card('Overall', `<div class="status ${htmlEscape(overall.overallStatus || 'UNKNOWN')}">${htmlEscape(overall.overallStatus || 'UNKNOWN')}</div>`),
    card('Open Issues', `<div>${htmlEscape(overall.openIssuesCount ?? 0)}</div>`),
    card('Template View', `<div>${htmlEscape(currentTemplateLabel())}</div>`),
    card('Recover Trend', `<div>${htmlEscape(trend)}</div><div class="muted">${htmlEscape(lastRecover)}</div><div class="muted">cooldown: ${htmlEscape(cooldown)}</div>`),
    card('Last Refresh', `<div class="muted">${htmlEscape(overall.lastRefreshTime || '-')}</div>`),
  ].join('');
}

function renderStatusWall(overall, lifecycle, services, artifacts) {
  const target = document.getElementById('status-wall');
  const visibleStageIds = visibleStageIdsForTemplate(selectedTemplate);
  const stageMap = new Map((lifecycle.stages || []).map((s) => [s.stageId, s]));
  const counts = {
    PASSED: 0,
    WARN: 0,
    FAILED: 0,
    SKIPPED: 0,
    NOT_STARTED: 0,
    RUNNING: 0,
  };
  visibleStageIds.forEach((id) => {
    const status = (stageMap.get(id)?.status || 'NOT_STARTED');
    if (Object.prototype.hasOwnProperty.call(counts, status)) {
      counts[status] += 1;
    }
  });
  const scored = counts.PASSED + counts.WARN + counts.FAILED;
  const healthScore = scored ? Math.round(((counts.PASSED + counts.WARN * 0.5) / scored) * 100) : 0;

  const serviceRows = services.services || [];
  const serviceHealthy = serviceRows.filter((x) => String(x.health || '').toLowerCase() === 'healthy').length;
  const serviceDegraded = serviceRows.filter((x) => String(x.health || '').toLowerCase() === 'degraded').length;
  const serviceUnhealthy = serviceRows.filter((x) => String(x.health || '').toLowerCase() === 'unhealthy').length;

  const bundles = artifacts.bundles || [];
  const reports = artifacts.reports || [];

  target.innerHTML = [
    wallCard('运行健康度', `${healthScore}%`, `visible stages ${visibleStageIds.length}`),
    wallCard('当前状态', overall.overallStatus || 'UNKNOWN', `issues ${overall.openIssuesCount || 0}`),
    wallCard('阶段分布', `${counts.PASSED}/${visibleStageIds.length}`, `warn ${counts.WARN} · fail ${counts.FAILED}`),
    wallCard('服务健康', `${serviceHealthy}`, `degraded ${serviceDegraded} · unhealthy ${serviceUnhealthy}`),
    wallCard('证据产物', `${reports.length}/${bundles.length}`, 'reports / bundles', 'small'),
    wallCard('模板视图', selectedTemplate ? selectedTemplate.mode : 'generic', selectedTemplate ? (selectedTemplate.ref || '-') : 'no template selected', 'small'),
  ].join('');
}

function parseSystemInfoMessage(msg) {
  const out = {};
  const osMatch = /(?:^|\s)os=([^\s]+)/i.exec(msg || '');
  const archMatch = /(?:^|\s)arch=([^\s]+)/i.exec(msg || '');
  const kernelTzMatch = /kernel=(.+?)\s+timezone=/i.exec(msg || '');
  const kernelOnlyMatch = /kernel=(.+)$/i.exec(msg || '');
  const tzMatch = /(?:^|\s)timezone=([^\s]+)/i.exec(msg || '');
  if (osMatch) out.os = osMatch[1];
  if (archMatch) out.arch = archMatch[1];
  if (kernelTzMatch) out.kernel = kernelTzMatch[1].trim();
  else if (kernelOnlyMatch) out.kernel = kernelOnlyMatch[1].trim();
  if (tzMatch) out.timezone = tzMatch[1];
  return out;
}

function extractServerInfo(lifecycle, services) {
  const info = {
    os: '-',
    arch: '-',
    kernel: '-',
    timezone: '-',
    host: window.location.hostname || '-',
  };

  const stageA = (lifecycle.stages || []).find((x) => x.stageId === 'A');
  (stageA?.metrics || []).forEach((m) => {
    const label = String(m.label || '').toLowerCase();
    if (label === 'os' && info.os === '-') info.os = String(m.value || '-');
    if (label === 'arch' && info.arch === '-') info.arch = String(m.value || '-');
  });

  const checks = [];
  (services.services || []).forEach((svc) => {
    (svc.checks || []).forEach((c) => checks.push(c));
  });

  const sysInfoCheck = checks.find((c) => String(c.checkId || '').includes('system_info'));
  if (sysInfoCheck && String(sysInfoCheck.message || '').trim() !== '') {
    const parsed = parseSystemInfoMessage(sysInfoCheck.message);
    info.os = parsed.os || info.os;
    info.arch = parsed.arch || info.arch;
    info.kernel = parsed.kernel || info.kernel;
    info.timezone = parsed.timezone || info.timezone;
  }

  return info;
}

function renderServerBasic(overall, lifecycle, services) {
  const target = document.getElementById('server-basic');
  const info = extractServerInfo(lifecycle, services);
  const items = [
    ['Host', info.host],
    ['OS', info.os],
    ['Arch', info.arch],
    ['Kernel', info.kernel],
    ['TZ', info.timezone],
    ['Refresh', overall.lastRefreshTime || '-'],
  ];
  target.innerHTML = items.map(([k, v]) => `<span class="kv"><b>${htmlEscape(k)}</b><span>${htmlEscape(v || '-')}</span></span>`).join('');
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

function visibleStageIdsForTemplate(template) {
  if (!template) return ['A', 'D', 'E', 'F'];
  const mode = String(template.mode || '').toLowerCase();
  if (mode === 'deploy') return ['A', 'B', 'C', 'D', 'E', 'F'];
  if (mode === 'manage') return ['A', 'B', 'D', 'E', 'F'];
  return ['A', 'D', 'E', 'F'];
}

function renderLifecycleViewNote(template, lifecycle) {
  const title = document.getElementById('lifecycle-title');
  const note = document.getElementById('lifecycle-note');
  const all = ['A', 'B', 'C', 'D', 'E', 'F'];
  const visible = visibleStageIdsForTemplate(template);
  const hidden = all.filter((x) => !visible.includes(x));
  if (template) {
    title.textContent = `Lifecycle (A-F) · ${template.ref}`;
    note.textContent = hidden.length
      ? `当前模板视图模式：${template.mode}，展示阶段 ${visible.join(',')}，隐藏 ${hidden.join(',')}。`
      : `当前模板视图模式：${template.mode}，展示全部 A-F 阶段。`;
    return;
  }
  title.textContent = 'Lifecycle (A-F) · Generic Baseline';
  note.textContent = `未选择模板，按基础通用能力展示阶段 ${visible.join(',')}。`;
}

function renderStages(lifecycle, artifacts) {
  const stageIds = visibleStageIdsForTemplate(selectedTemplate);
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
  target.innerHTML = stageIds.map((id) => {
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
    const summaryHtml = stageSummaryLine(s.summary);
    return card(`${s.stageId} ${s.name}`, `
      <div class="status ${htmlEscape(s.status)}">${htmlEscape(s.status)}</div>
      <div class="muted">${htmlEscape(s.lastRunTime || '-')}</div>
      <div class="muted">issues: ${(s.issues || []).length}</div>
      ${summaryHtml}
      ${triggerHtml}
      ${recoverLinksHtml}
    `);
  }).join('');
}

function stageMetricValue(stage, label) {
  const metric = (stage.metrics || []).find((m) => m.label === label);
  return metric ? metric.value : '';
}

function stageSummaryLine(summary) {
  if (!summary) return '';
  const total = Number(summary.total || 0);
  const pass = Number(summary.pass || 0);
  const warn = Number(summary.warn || 0);
  const fail = Number(summary.fail || 0);
  const skip = Number(summary.skip || 0);
  return `<div class="muted">steps: pass ${pass} · warn ${warn} · fail ${fail} · skip ${skip} (total ${total})</div>`;
}

function renderServices(services) {
  const target = document.getElementById('services');
  const rows = (services.services || []).map((svc) => {
    const checks = (svc.checks || []).map((c) => `${c.checkId}:${c.result}`).join(', ');
    return `<li><strong>${htmlEscape(svc.serviceId || '-')}</strong> (${htmlEscape(svc.health || 'unknown')}) - ${htmlEscape(checks || 'no checks')}</li>`;
  });
  target.innerHTML = rows.length ? `<ul>${rows.join('')}</ul>` : '<div class="muted">no services yet</div>';
}

function findSelectedTemplate(catalog) {
  const items = (catalog && Array.isArray(catalog.templates)) ? catalog.templates : [];
  if (!selectedTemplateRef) return null;
  const ref = String(selectedTemplateRef).trim();
  const found = items.find((x) => x.ref === ref || x.templateId === ref);
  if (!found) return null;
  return found;
}

function renderTemplateSelector(catalog, overall) {
  const select = document.getElementById('template-select');
  const hint = document.getElementById('template-hint');
  const items = (catalog && Array.isArray(catalog.templates)) ? catalog.templates : [];

  const options = ['<option value="">基础通用能力（不选模板）</option>'];
  items
    .slice()
    .sort((a, b) => String(a.ref || '').localeCompare(String(b.ref || '')))
    .forEach((item) => {
      const label = `${item.ref} (${item.mode}/${item.serviceScope})`;
      options.push(`<option value="${htmlEscape(item.ref)}">${htmlEscape(label)}</option>`);
    });
  select.innerHTML = options.join('');

  const found = findSelectedTemplate(catalog);
  if (!found) {
    selectedTemplateRef = '';
    selectedTemplate = null;
  } else {
    selectedTemplate = found;
  }
  select.value = selectedTemplateRef || '';

  if (selectedTemplate) {
    hint.textContent = `当前模板：${selectedTemplate.ref} · mode=${selectedTemplate.mode} · scope=${selectedTemplate.serviceScope}`;
  } else {
    const active = (overall.activeTemplates || []).join(', ') || '-';
    hint.textContent = `当前为基础通用能力视图。activeTemplates: ${active}`;
  }

  select.onchange = () => {
    selectedTemplateRef = String(select.value || '');
    syncTemplateViewToURL(selectedTemplateRef);
    renderAll();
  };
}

function renderTemplates(catalog) {
  const target = document.getElementById('templates');
  const items = (catalog && Array.isArray(catalog.templates)) ? catalog.templates : [];
  if (!items.length) {
    target.innerHTML = '<div class="muted">template catalog not found yet. Run `opskit status` to refresh state/templates.json.</div>';
    return;
  }
  const filtered = filterTemplates(items, templateFilter);
  if (!filtered.length) {
    target.innerHTML = '<div class="muted">no templates match current filter</div>';
    return;
  }
  const groups = groupTemplates(filtered);
  const groupOrder = ['manage|single-service', 'manage|multi-service', 'deploy|single-service', 'deploy|multi-service'];
  const rows = groupOrder
    .filter((key) => groups.has(key))
    .map((key) => {
      const [mode, scope] = key.split('|');
      const list = groups.get(key) || [];
      const title = `${mode} / ${scope}`;
      const lines = list.map((item) => {
        const aliases = (item.aliases || []).length ? ` aliases: ${(item.aliases || []).join(',')}` : '';
        const tags = (item.tags || []).map((x) => `<span class="tag">${htmlEscape(x)}</span>`).join(' ');
        return `<li><strong>${htmlEscape(item.ref || '-')}</strong> <span class="muted">id=${htmlEscape(item.templateId || '-')} source=${htmlEscape(item.source || '-')}${htmlEscape(aliases)}</span><div class="tag-row">${tags}</div></li>`;
      });
      return `<h3>${htmlEscape(title)}</h3><ul>${lines.join('')}</ul>`;
    });
  target.innerHTML = rows.join('');
}

function groupTemplates(items) {
  const groups = new Map();
  items.forEach((item) => {
    const mode = String(item.mode || 'unknown').toLowerCase();
    const scope = String(item.serviceScope || 'single-service').toLowerCase();
    const key = `${mode}|${scope}`;
    if (!groups.has(key)) {
      groups.set(key, []);
    }
    groups.get(key).push(item);
  });
  groups.forEach((list) => {
    list.sort((a, b) => String(a.ref || '').localeCompare(String(b.ref || '')));
  });
  return groups;
}

function filterTemplates(items, filter) {
  if (filter === 'all') return items;
  return items.filter((item) => {
    const mode = String(item.mode || '').toLowerCase();
    const scope = String(item.serviceScope || '').toLowerCase();
    const tags = (item.tags || []).map((x) => String(x).toLowerCase());
    return mode === filter || scope === filter || tags.includes(filter);
  });
}

function renderTemplateFilters(catalog) {
  const target = document.getElementById('template-filters');
  const items = (catalog && Array.isArray(catalog.templates)) ? catalog.templates : [];
  if (!items.length) {
    target.innerHTML = '';
    return;
  }
  const defs = [
    { id: 'all', label: 'All' },
    { id: 'manage', label: 'Manage' },
    { id: 'deploy', label: 'Deploy' },
    { id: 'single-service', label: 'Single Service' },
    { id: 'multi-service', label: 'Multi Service' },
    { id: 'demo', label: 'Demo' },
    { id: 'builtin', label: 'Builtin' },
  ];
  const count = (id) => filterTemplates(items, id).length;
  target.innerHTML = defs.map((d) => {
    const active = d.id === templateFilter ? 'active' : '';
    return `<button class="filter-btn ${active}" data-template-filter="${htmlEscape(d.id)}">${htmlEscape(d.label)} (${count(d.id)})</button>`;
  }).join('');
  target.querySelectorAll('.filter-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      templateFilter = btn.dataset.templateFilter || 'all';
      syncTemplateFilterToURL(templateFilter);
      renderTemplateFilters(catalog);
      renderTemplates(catalog);
    });
  });
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
      renderArtifacts(artifacts, appState ? appState.summary : null);
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

function allowedTemplateFilter(id) {
  return ['all', 'manage', 'deploy', 'single-service', 'multi-service', 'demo', 'builtin'].includes(id);
}

function parseTemplateFilterFromURL() {
  const params = new URLSearchParams(window.location.search || '');
  const raw = (params.get('templateFilter') || 'all').trim().toLowerCase();
  return allowedTemplateFilter(raw) ? raw : 'all';
}

function syncTemplateFilterToURL(filter) {
  const url = new URL(window.location.href);
  if (!allowedTemplateFilter(filter) || filter === 'all') {
    url.searchParams.delete('templateFilter');
  } else {
    url.searchParams.set('templateFilter', filter);
  }
  window.history.replaceState(null, '', `${url.pathname}${url.search}${url.hash}`);
}

function parseTemplateViewFromURL() {
  const params = new URLSearchParams(window.location.search || '');
  return String(params.get('viewTemplate') || '').trim();
}

function syncTemplateViewToURL(ref) {
  const url = new URL(window.location.href);
  if (!ref) {
    url.searchParams.delete('viewTemplate');
  } else {
    url.searchParams.set('viewTemplate', ref);
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

function startRecoverCountdown() {
  if (recoverCountdownTimer) {
    clearInterval(recoverCountdownTimer);
    recoverCountdownTimer = null;
  }
  if (!appState) return;
  const recover = appState.overall.recoverSummary || {};
  if (!recover.circuitOpen || !isCooldownOpen(recover.cooldownUntil)) {
    return;
  }
  recoverCountdownTimer = setInterval(() => {
    if (!appState) return;
    renderRecoverAlert(appState.overall);
    renderOverview(appState.overall);
    renderStatusWall(appState.overall, appState.lifecycle, appState.services, appState.artifacts);
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
  const latestAcceptConsistency = latestBy(
    (x) => (x.path || '').includes('acceptance-consistency-') || String(x.id || '').includes('consistency'),
    reports,
  ) || consistencyFromAcceptanceBundle(latestAccept);
  const latestCollect = latestBy((x) => (x.path || '').includes('collect-') || String(x.id || '').includes('collect'), bundles);
  const latestHandover = latestBy((x) => (x.path || '').includes('handover-') || String(x.id || '').includes('handover'), bundles);
  const summaryCard = summary
    ? card('Generic E2E', `<a class="artifact-link" href="../summary.json" target="_blank" rel="noopener">summary.json</a><div class="muted">overall: ${htmlEscape(summary.overall || 'unknown')} · failed ${htmlEscape(summary.failedSteps ?? 0)}/${htmlEscape(summary.totalSteps ?? 0)}</div>`)
    : card('Generic E2E', '<span class="muted">summary.json not found</span>');
  target.innerHTML = [
    card('Counts', `<div>reports: ${reports.length}</div><div>bundles: ${bundles.length}</div>`),
    card('Latest Acceptance', artifactLink('bundle', latestAccept)),
    card('Latest Acceptance Consistency', artifactLink('report', latestAcceptConsistency)),
    card('Latest Recover Collect', artifactLink('bundle', latestCollect)),
    card('Latest Handover', artifactLink('bundle', latestHandover)),
    summaryCard,
  ].join('');
}

function consistencyFromAcceptanceBundle(artifact) {
  const path = String(artifact?.path || '');
  const match = path.match(/acceptance-(\d{8}-\d{6})\.tar\.gz$/);
  if (!match) return null;
  return {
    id: `acceptance-consistency-${match[1]}`,
    path: `evidence/acceptance-consistency-${match[1]}.json`,
  };
}

function renderAll() {
  if (!appState) return;
  const { overall, lifecycle, services, artifacts, summary, templatesCatalog } = appState;
  selectedTemplate = findSelectedTemplate(templatesCatalog);

  const headline = document.getElementById('headline');
  const stageView = visibleStageIdsForTemplate(selectedTemplate).join(',');
  headline.textContent = `Overall ${overall.overallStatus || 'UNKNOWN'} · refreshed ${overall.lastRefreshTime || '-'} · view ${stageView}`;

  renderTemplateSelector(templatesCatalog, overall);
  renderServerBasic(overall, lifecycle, services);
  renderRecoverAlert(overall);
  renderStatusWall(overall, lifecycle, services, artifacts);
  renderLifecycleViewNote(selectedTemplate, lifecycle);
  renderStages(lifecycle, artifacts);
  renderOverview(overall);
  renderServices(services);
  renderTemplateFilters(templatesCatalog);
  renderTemplates(templatesCatalog);
  renderArtifactHighlights(artifacts, summary);
  renderArtifactFilters(artifacts);
  renderArtifacts(artifacts, summary);
  startRecoverCountdown();
}

async function boot() {
  const headline = document.getElementById('headline');
  try {
    const [overall, lifecycle, services, artifacts, summary, templatesCatalog] = await Promise.all([
      loadJson('overall'),
      loadJson('lifecycle'),
      loadJson('services'),
      loadJson('artifacts'),
      loadOptionalJson('../summary.json'),
      loadOptionalJson(`${statePrefix}/templates.json`),
    ]);

    appState = {
      overall,
      lifecycle,
      services,
      artifacts,
      summary,
      templatesCatalog,
    };
    renderAll();
  } catch (err) {
    headline.textContent = `Failed to load state JSON: ${err.message}`;
  }
}

boot();
