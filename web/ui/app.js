const statePrefix = '../state';
let artifactFilter = parseFilterFromURL();
let templateFilter = parseTemplateFilterFromURL();
let stageFilter = parseStageFilterFromURL();
let serviceFilter = parseServiceFilterFromURL();
let selectedTemplateRef = parseTemplateViewFromURL();
let selectedTemplate = null;
let recoverCountdownTimer = null;
let appState = null;
let autoRefreshTimer = null;
let autoRefreshEnabled = readBoolSetting('opskit.ui.autoRefresh', true);
let autoRefreshIntervalSec = readIntSetting('opskit.ui.autoRefreshSec', 15, [5, 10, 15, 30, 60]);
let inFlightReload = false;

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

function readBoolSetting(key, fallback) {
  try {
    const raw = window.localStorage.getItem(key);
    if (raw === null) return fallback;
    return raw === '1';
  } catch (_) {
    return fallback;
  }
}

function readIntSetting(key, fallback, allowed) {
  try {
    const raw = Number(window.localStorage.getItem(key));
    if (!Number.isInteger(raw)) return fallback;
    return allowed.includes(raw) ? raw : fallback;
  } catch (_) {
    return fallback;
  }
}

function writeSetting(key, value) {
  try {
    window.localStorage.setItem(key, String(value));
  } catch (_) {
    // ignore storage write failures
  }
}

function card(title, bodyHtml) {
  return `<article class="card"><h3>${htmlEscape(title)}</h3>${bodyHtml}</article>`;
}

function wallCard(title, value, sub, extraClass = '') {
  return `<article class="wall-card ${htmlEscape(extraClass)}"><h3>${htmlEscape(title)}</h3><div class="wall-value ${htmlEscape(extraClass)}">${htmlEscape(value)}</div><div class="wall-sub">${htmlEscape(sub)}</div></article>`;
}

function wallActionCard(action, title, value, sub, extraClass = '') {
  return `<article class="wall-card clickable ${htmlEscape(extraClass)}" data-action="${htmlEscape(action)}"><h3>${htmlEscape(title)}</h3><div class="wall-value ${htmlEscape(extraClass)}">${htmlEscape(value)}</div><div class="wall-sub">${htmlEscape(sub)}</div></article>`;
}

function currentTemplateLabel() {
  if (!selectedTemplate) {
    return 'generic-baseline';
  }
  return `${selectedTemplate.ref} (${selectedTemplate.mode}/${selectedTemplate.serviceScope})`;
}

function templateMatchesGroup(item, groupId) {
  if (!item) return false;
  const mode = String(item.mode || '').toLowerCase();
  const scope = String(item.serviceScope || '').toLowerCase();
  const tags = (item.tags || []).map((x) => String(x).toLowerCase());
  if (groupId === 'manage' || groupId === 'deploy') return mode === groupId;
  if (groupId === 'single-service' || groupId === 'multi-service') return scope === groupId;
  if (groupId === 'demo' || groupId === 'builtin') return tags.includes(groupId);
  return false;
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
    wallActionCard('wall:stage-all', '运行健康度', `${healthScore}%`, `visible stages ${visibleStageIds.length}`),
    wallCard('当前状态', overall.overallStatus || 'UNKNOWN', `issues ${overall.openIssuesCount || 0}`),
    wallActionCard('wall:stage-problem', '阶段分布', `${counts.PASSED}/${visibleStageIds.length}`, `warn ${counts.WARN} · fail ${counts.FAILED}`),
    wallActionCard('wall:service-degraded', '服务健康', `${serviceHealthy}`, `degraded ${serviceDegraded} · unhealthy ${serviceUnhealthy}`),
    wallActionCard('wall:artifact-all', '证据产物', `${reports.length}/${bundles.length}`, 'reports / bundles', 'small'),
    wallCard('模板视图', selectedTemplate ? selectedTemplate.mode : 'generic', selectedTemplate ? (selectedTemplate.ref || '-') : 'no template selected', 'small'),
  ].join('');
}

function severityRank(input) {
  const sev = String(input || '').toLowerCase();
  if (sev === 'fail' || sev === 'error') return 3;
  if (sev === 'warn' || sev === 'warning') return 2;
  return 1;
}

function renderDashboardPanels(overall, lifecycle, services, artifacts, templatesCatalog) {
  const target = document.getElementById('dashboard-grid');
  if (!target) return;

  const visibleStageIds = visibleStageIdsForTemplate(selectedTemplate);
  const stageMap = new Map((lifecycle.stages || []).map((x) => [x.stageId, x]));
  const stageRows = visibleStageIds.map((id) => {
    const row = stageMap.get(id) || { stageId: id, name: '-', status: 'NOT_STARTED' };
    return {
      stageId: id,
      name: row.name || '-',
      status: row.status || 'NOT_STARTED',
      issues: row.issues || [],
    };
  });
  const stageCounts = { PASSED: 0, WARN: 0, FAILED: 0, RUNNING: 0, NOT_STARTED: 0, SKIPPED: 0 };
  stageRows.forEach((row) => {
    if (!Object.prototype.hasOwnProperty.call(stageCounts, row.status)) {
      stageCounts.NOT_STARTED += 1;
      return;
    }
    stageCounts[row.status] += 1;
  });
  const totalStage = Math.max(stageRows.length, 1);
  const progressSegments = ['PASSED', 'WARN', 'FAILED', 'RUNNING', 'NOT_STARTED', 'SKIPPED']
    .map((status) => {
      const count = stageCounts[status];
      if (!count) return '';
      const width = Math.max(2, Math.round((count / totalStage) * 100));
      return `<span class="progress-seg ${htmlEscape(status)}" style="width:${width}%"></span>`;
    })
    .join('');
  const stageList = stageRows
    .map((row) => `<div class="stage-mini"><span class="name">${htmlEscape(row.stageId)} ${htmlEscape(row.name)}</span><span class="status ${htmlEscape(row.status)}">${htmlEscape(row.status)}</span></div>`)
    .join('');

  const issueRows = stageRows
    .flatMap((row) => (row.issues || []).map((issue) => ({
      stageId: row.stageId,
      severity: String(issue.severity || 'info').toLowerCase(),
      message: issue.message || issue.id || '-',
    })))
    .sort((a, b) => severityRank(b.severity) - severityRank(a.severity))
    .slice(0, 8);
  const issueList = issueRows.length
    ? `<ul class="dash-list">${issueRows.map((row) => `<li><span class="severity-badge ${htmlEscape(row.severity)}">${htmlEscape(row.severity.toUpperCase())}</span>[${htmlEscape(row.stageId)}] ${htmlEscape(row.message)}</li>`).join('')}</ul>`
    : '<div class="muted">当前视图无告警/失败问题</div>';

  const serviceRows = services.services || [];
  const serviceList = serviceRows.length
    ? `<ul class="dash-list">${serviceRows.map((svc) => `<li><strong>${htmlEscape(svc.serviceId || '-')}</strong> <span class="status ${htmlEscape(String(svc.health || '').toUpperCase())}">${htmlEscape((svc.health || 'unknown').toUpperCase())}</span></li>`).join('')}</ul>`
    : '<div class="muted">暂无服务数据</div>';

  const bundles = artifacts.bundles || [];
  const reports = artifacts.reports || [];
  const latestAcceptance = latestBy((x) => artifactKind(x) === 'acceptance', bundles);
  const latestCollect = latestBy((x) => artifactKind(x) === 'recover', bundles);
  const latestHandover = latestBy((x) => artifactKind(x) === 'handover', bundles);
  const latestConsistency = latestBy((x) => String(x.id || '').includes('consistency') || String(x.path || '').includes('consistency'), reports);

  const templateItems = (templatesCatalog && Array.isArray(templatesCatalog.templates)) ? templatesCatalog.templates : [];
  const manageCount = templateItems.filter((x) => String(x.mode || '').toLowerCase() === 'manage').length;
  const deployCount = templateItems.filter((x) => String(x.mode || '').toLowerCase() === 'deploy').length;
  const scopeMulti = templateItems.filter((x) => String(x.serviceScope || '').toLowerCase() === 'multi-service').length;
  const scopeSingle = templateItems.filter((x) => String(x.serviceScope || '').toLowerCase() === 'single-service').length;

  target.innerHTML = [
    `<article class="dash-panel clickable" data-action="dash:stage-all"><h3>阶段进度</h3><div class="progress-summary">${progressSegments || '<span class="progress-seg NOT_STARTED" style="width:100%"></span>'}</div><div class="muted">pass ${stageCounts.PASSED} · warn ${stageCounts.WARN} · fail ${stageCounts.FAILED} · running ${stageCounts.RUNNING}</div>${stageList}</article>`,
    `<article class="dash-panel clickable" data-action="dash:stage-problem"><h3>问题聚合（Top 8）</h3>${issueList}</article>`,
    `<article class="dash-panel clickable" data-action="dash:service-degraded"><h3>服务矩阵</h3>${serviceList}</article>`,
    `<article class="dash-panel clickable" data-action="dash:artifact-acceptance"><h3>证据通道</h3><ul class="dash-list"><li>reports: ${reports.length} · bundles: ${bundles.length}</li><li>acceptance: ${htmlEscape(latestAcceptance ? (latestAcceptance.id || latestAcceptance.path) : '-')}</li><li>recover collect: ${htmlEscape(latestCollect ? (latestCollect.id || latestCollect.path) : '-')}</li><li>handover: ${htmlEscape(latestHandover ? (latestHandover.id || latestHandover.path) : '-')}</li><li>consistency: ${htmlEscape(latestConsistency ? (latestConsistency.id || latestConsistency.path) : '-')}</li></ul></article>`,
    `<article class="dash-panel clickable" data-action="dash:template-all"><h3>模板库存</h3><ul class="dash-list"><li>catalog total: ${templateItems.length}</li><li>manage/deploy: ${manageCount}/${deployCount}</li><li>single/multi: ${scopeSingle}/${scopeMulti}</li><li>view: ${htmlEscape(currentTemplateLabel())}</li></ul></article>`,
  ].join('');
}

function scrollToSection(id) {
  const el = document.getElementById(id);
  if (!el) return;
  try {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  } catch (_) {
    el.scrollIntoView();
  }
}

function applyDashboardAction(action) {
  switch (action) {
    case 'wall:stage-all':
    case 'dash:stage-all':
      stageFilter = 'all';
      syncStageFilterToURL(stageFilter);
      scrollToSection('stages');
      break;
    case 'wall:stage-problem':
    case 'dash:stage-problem':
      stageFilter = 'problem';
      syncStageFilterToURL(stageFilter);
      scrollToSection('stages');
      break;
    case 'wall:service-degraded':
    case 'dash:service-degraded':
      serviceFilter = 'degraded';
      syncServiceFilterToURL(serviceFilter);
      scrollToSection('services');
      break;
    case 'wall:artifact-all':
      artifactFilter = 'all';
      syncFilterToURL(artifactFilter);
      scrollToSection('artifacts');
      break;
    case 'dash:artifact-acceptance':
      artifactFilter = 'acceptance';
      syncFilterToURL(artifactFilter);
      scrollToSection('artifacts');
      break;
    case 'dash:template-all':
      templateFilter = 'all';
      syncTemplateFilterToURL(templateFilter);
      scrollToSection('templates');
      break;
    default:
      return;
  }
  renderAll();
}

function wireDashboardActions() {
  const actionNodes = document.querySelectorAll('[data-action]');
  actionNodes.forEach((node) => {
    node.addEventListener('click', () => {
      const action = node.getAttribute('data-action') || '';
      applyDashboardAction(action);
    });
  });
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

function stageStatusMatch(status, filter) {
  const normalized = String(status || 'NOT_STARTED').toUpperCase();
  if (filter === 'all') return true;
  if (filter === 'problem') return normalized === 'WARN' || normalized === 'FAILED';
  return normalized.toLowerCase() === filter;
}

function renderLifecycleViewNote(template, lifecycle) {
  const title = document.getElementById('lifecycle-title');
  const note = document.getElementById('lifecycle-note');
  const all = ['A', 'B', 'C', 'D', 'E', 'F'];
  const visible = visibleStageIdsForTemplate(template);
  const hidden = all.filter((x) => !visible.includes(x));
  const stageFilterLabel = stageFilter === 'all' ? 'all' : stageFilter;
  if (template) {
    title.textContent = `Lifecycle (A-F) · ${template.ref}`;
    note.textContent = hidden.length
      ? `当前模板视图模式：${template.mode}，展示阶段 ${visible.join(',')}，隐藏 ${hidden.join(',')}。stage filter=${stageFilterLabel}`
      : `当前模板视图模式：${template.mode}，展示全部 A-F 阶段。stage filter=${stageFilterLabel}`;
    return;
  }
  title.textContent = 'Lifecycle (A-F) · Generic Baseline';
  note.textContent = `未选择模板，按基础通用能力展示阶段 ${visible.join(',')}。stage filter=${stageFilterLabel}`;
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
  const filteredStageIds = stageIds.filter((id) => {
    const s = (lifecycle.stages || []).find((x) => x.stageId === id);
    return stageStatusMatch(s?.status || 'NOT_STARTED', stageFilter);
  });
  if (!filteredStageIds.length) {
    target.innerHTML = '<div class="card"><h3>Stages</h3><div class="muted">no stages match current stage filter</div></div>';
    return;
  }

  target.innerHTML = filteredStageIds.map((id) => {
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

function renderStageFilters(lifecycle) {
  const target = document.getElementById('stage-filters');
  if (!target) return;
  const stageIds = visibleStageIdsForTemplate(selectedTemplate);
  const stages = stageIds.map((id) => (lifecycle.stages || []).find((x) => x.stageId === id) || { status: 'NOT_STARTED' });
  const defs = [
    { id: 'all', label: 'All' },
    { id: 'problem', label: 'Problem' },
    { id: 'passed', label: 'Passed' },
    { id: 'warn', label: 'Warn' },
    { id: 'failed', label: 'Failed' },
    { id: 'running', label: 'Running' },
    { id: 'not_started', label: 'Not Started' },
  ];
  const count = (id) => stages.filter((s) => stageStatusMatch(s.status || 'NOT_STARTED', id)).length;
  target.innerHTML = defs.map((d) => {
    const active = d.id === stageFilter ? 'active' : '';
    return `<button class="filter-btn ${active}" data-stage-filter="${htmlEscape(d.id)}">${htmlEscape(d.label)} (${count(d.id)})</button>`;
  }).join('');
  target.querySelectorAll('.filter-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      stageFilter = btn.dataset.stageFilter || 'all';
      syncStageFilterToURL(stageFilter);
      renderStageFilters(lifecycle);
      renderLifecycleViewNote(selectedTemplate, lifecycle);
      renderStages(lifecycle, appState ? appState.artifacts : { reports: [], bundles: [] });
    });
  });
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
  const rows = (services.services || [])
    .filter((svc) => {
      if (serviceFilter === 'all') return true;
      return String(svc.health || '').toLowerCase() === serviceFilter;
    })
    .map((svc) => {
    const checks = (svc.checks || []).map((c) => `${c.checkId}:${c.result}`).join(', ');
    return `<li><strong>${htmlEscape(svc.serviceId || '-')}</strong> (${htmlEscape(svc.health || 'unknown')}) - ${htmlEscape(checks || 'no checks')}</li>`;
  });
  target.innerHTML = rows.length ? `<ul>${rows.join('')}</ul>` : '<div class="muted">no services match current service filter</div>';
}

function renderServiceFilters(services) {
  const target = document.getElementById('service-filters');
  if (!target) return;
  const defs = [
    { id: 'all', label: 'All' },
    { id: 'healthy', label: 'Healthy' },
    { id: 'degraded', label: 'Degraded' },
    { id: 'unhealthy', label: 'Unhealthy' },
  ];
  const rows = services.services || [];
  const count = (id) => {
    if (id === 'all') return rows.length;
    return rows.filter((x) => String(x.health || '').toLowerCase() === id).length;
  };
  target.innerHTML = defs.map((d) => {
    const active = d.id === serviceFilter ? 'active' : '';
    return `<button class="filter-btn ${active}" data-service-filter="${htmlEscape(d.id)}">${htmlEscape(d.label)} (${count(d.id)})</button>`;
  }).join('');
  target.querySelectorAll('.filter-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      serviceFilter = btn.dataset.serviceFilter || 'all';
      syncServiceFilterToURL(serviceFilter);
      renderServiceFilters(services);
      renderServices(services);
    });
  });
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
  const clearBtn = document.getElementById('template-clear');
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
  if (clearBtn) {
    clearBtn.disabled = !selectedTemplateRef;
    clearBtn.onclick = () => {
      selectedTemplateRef = '';
      selectedTemplate = null;
      syncTemplateViewToURL('');
      renderAll();
    };
  }

  select.onchange = () => {
    selectedTemplateRef = String(select.value || '');
    syncTemplateViewToURL(selectedTemplateRef);
    renderAll();
  };
}

function pickTemplateForGroup(items, groupId) {
  if (!Array.isArray(items) || !items.length) return null;
  const sorted = items.slice().sort((a, b) => String(a.ref || '').localeCompare(String(b.ref || '')));
  if (selectedTemplate && templateMatchesGroup(selectedTemplate, groupId)) {
    return selectedTemplate;
  }
  return sorted.find((x) => templateMatchesGroup(x, groupId)) || null;
}

function renderTemplateGroupSwitch(catalog) {
  const target = document.getElementById('template-group-switch');
  if (!target) return;
  const items = (catalog && Array.isArray(catalog.templates)) ? catalog.templates : [];
  const defs = [
    { id: 'baseline', label: '基础通用' },
    { id: 'manage', label: 'Manage' },
    { id: 'deploy', label: 'Deploy' },
    { id: 'single-service', label: '单服务' },
    { id: 'multi-service', label: '多服务' },
    { id: 'demo', label: 'Demo' },
    { id: 'builtin', label: 'Builtin' },
  ];
  const count = (id) => {
    if (id === 'baseline') return 1;
    return items.filter((x) => templateMatchesGroup(x, id)).length;
  };
  const isActive = (id) => {
    if (id === 'baseline') return !selectedTemplate;
    return !!selectedTemplate && templateMatchesGroup(selectedTemplate, id);
  };

  target.innerHTML = defs.map((d) => {
    const active = isActive(d.id) ? 'active' : '';
    return `<button class="filter-btn ${active}" data-template-group="${htmlEscape(d.id)}">${htmlEscape(d.label)} (${count(d.id)})</button>`;
  }).join('');

  target.querySelectorAll('.filter-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      const groupId = btn.dataset.templateGroup || 'baseline';
      if (groupId === 'baseline') {
        selectedTemplateRef = '';
        selectedTemplate = null;
        templateFilter = 'all';
        syncTemplateViewToURL('');
        syncTemplateFilterToURL('all');
        renderAll();
        return;
      }
      const picked = pickTemplateForGroup(items, groupId);
      if (!picked) return;
      selectedTemplateRef = picked.ref || picked.templateId || '';
      selectedTemplate = picked;
      syncTemplateViewToURL(selectedTemplateRef);
      if (allowedTemplateFilter(groupId)) {
        templateFilter = groupId;
        syncTemplateFilterToURL(templateFilter);
      }
      renderAll();
    });
  });
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

function allowedStageFilter(id) {
  return ['all', 'problem', 'passed', 'warn', 'failed', 'running', 'not_started'].includes(id);
}

function parseStageFilterFromURL() {
  const params = new URLSearchParams(window.location.search || '');
  const raw = (params.get('stageFilter') || 'all').trim().toLowerCase();
  return allowedStageFilter(raw) ? raw : 'all';
}

function syncStageFilterToURL(filter) {
  const url = new URL(window.location.href);
  if (!allowedStageFilter(filter) || filter === 'all') {
    url.searchParams.delete('stageFilter');
  } else {
    url.searchParams.set('stageFilter', filter);
  }
  window.history.replaceState(null, '', `${url.pathname}${url.search}${url.hash}`);
}

function allowedServiceFilter(id) {
  return ['all', 'healthy', 'degraded', 'unhealthy'].includes(id);
}

function parseServiceFilterFromURL() {
  const params = new URLSearchParams(window.location.search || '');
  const raw = (params.get('serviceFilter') || 'all').trim().toLowerCase();
  return allowedServiceFilter(raw) ? raw : 'all';
}

function syncServiceFilterToURL(filter) {
  const url = new URL(window.location.href);
  if (!allowedServiceFilter(filter) || filter === 'all') {
    url.searchParams.delete('serviceFilter');
  } else {
    url.searchParams.set('serviceFilter', filter);
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

function renderRefreshHint(overall) {
  const hint = document.getElementById('refresh-hint');
  if (!hint) return;
  const base = overall && overall.lastRefreshTime ? `上次刷新: ${overall.lastRefreshTime}` : '上次刷新: -';
  if (!autoRefreshEnabled) {
    hint.textContent = `${base} · 自动刷新已关闭`;
    return;
  }
  hint.textContent = `${base} · 自动刷新 ${autoRefreshIntervalSec}s`;
}

function renderStatusRefreshHint(webRuntime) {
  const hint = document.getElementById('status-refresh-hint');
  if (!hint) return;
  if (!webRuntime || typeof webRuntime !== 'object') {
    hint.textContent = '后台状态刷新: unknown（未检测到 web_runtime.json）';
    return;
  }
  const enabled = !!webRuntime.statusAutoRefreshEnabled;
  const interval = Number(webRuntime.statusIntervalSeconds || 0);
  const lastTime = String(webRuntime.lastStatusRefreshTime || '-');
  const lastTrigger = String(webRuntime.lastStatusRefreshTrigger || '-');
  const lastExit = Number.isInteger(Number(webRuntime.lastStatusRefreshExit))
    ? String(webRuntime.lastStatusRefreshExit)
    : '-';
  const lastError = String(webRuntime.lastStatusRefreshError || '');
  if (!enabled) {
    hint.textContent = '后台状态刷新: 未启用（可通过 `opskit web --status-interval 15s` 启用）';
    return;
  }
  const errorPart = lastError ? ` · error=${lastError}` : '';
  hint.textContent = `后台状态刷新: 已启用(${interval || '?'}s) · last=${lastTime} · trigger=${lastTrigger} · exit=${lastExit}${errorPart}`;
}

function applyAutoRefreshTimer() {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer);
    autoRefreshTimer = null;
  }
  if (!autoRefreshEnabled) return;
  autoRefreshTimer = setInterval(() => {
    reloadState(true);
  }, autoRefreshIntervalSec * 1000);
}

function initRefreshControls() {
  const autoBox = document.getElementById('auto-refresh');
  const intervalSelect = document.getElementById('refresh-interval');
  const refreshNow = document.getElementById('refresh-now');
  if (!autoBox || !intervalSelect || !refreshNow) return;

  autoBox.checked = autoRefreshEnabled;
  intervalSelect.value = String(autoRefreshIntervalSec);
  intervalSelect.disabled = !autoRefreshEnabled;

  autoBox.onchange = () => {
    autoRefreshEnabled = !!autoBox.checked;
    intervalSelect.disabled = !autoRefreshEnabled;
    writeSetting('opskit.ui.autoRefresh', autoRefreshEnabled ? '1' : '0');
    applyAutoRefreshTimer();
    if (appState) {
      renderRefreshHint(appState.overall);
      renderStatusRefreshHint(appState.webRuntime);
    }
  };

  intervalSelect.onchange = () => {
    const raw = Number(intervalSelect.value);
    if (!Number.isInteger(raw) || ![5, 10, 15, 30, 60].includes(raw)) return;
    autoRefreshIntervalSec = raw;
    writeSetting('opskit.ui.autoRefreshSec', raw);
    applyAutoRefreshTimer();
    if (appState) {
      renderRefreshHint(appState.overall);
      renderStatusRefreshHint(appState.webRuntime);
    }
  };

  refreshNow.onclick = () => {
    reloadState(false);
  };
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

async function reloadState(silent) {
  if (inFlightReload) return;
  inFlightReload = true;
  const headline = document.getElementById('headline');
  try {
    const [overall, lifecycle, services, artifacts, summary, templatesCatalog, webRuntime] = await Promise.all([
      loadJson('overall'),
      loadJson('lifecycle'),
      loadJson('services'),
      loadJson('artifacts'),
      loadOptionalJson('../summary.json'),
      loadOptionalJson(`${statePrefix}/templates.json`),
      loadOptionalJson(`${statePrefix}/web_runtime.json`),
    ]);
    appState = {
      overall,
      lifecycle,
      services,
      artifacts,
      summary,
      templatesCatalog,
      webRuntime,
    };
    renderAll();
  } catch (err) {
    if (!silent && headline) {
      headline.textContent = `Failed to load state JSON: ${err.message}`;
    }
  } finally {
    inFlightReload = false;
  }
}

function renderAll() {
  if (!appState) return;
  const {
    overall,
    lifecycle,
    services,
    artifacts,
    summary,
    templatesCatalog,
    webRuntime,
  } = appState;
  selectedTemplate = findSelectedTemplate(templatesCatalog);

  const headline = document.getElementById('headline');
  const stageView = visibleStageIdsForTemplate(selectedTemplate).join(',');
  headline.textContent = `Overall ${overall.overallStatus || 'UNKNOWN'} · refreshed ${overall.lastRefreshTime || '-'} · view ${stageView}`;

  renderTemplateSelector(templatesCatalog, overall);
  renderTemplateGroupSwitch(templatesCatalog);
  renderServerBasic(overall, lifecycle, services);
  renderRecoverAlert(overall);
  renderStatusWall(overall, lifecycle, services, artifacts);
  renderDashboardPanels(overall, lifecycle, services, artifacts, templatesCatalog);
  wireDashboardActions();
  renderLifecycleViewNote(selectedTemplate, lifecycle);
  renderStageFilters(lifecycle);
  renderStages(lifecycle, artifacts);
  renderOverview(overall);
  renderServiceFilters(services);
  renderRefreshHint(overall);
  renderStatusRefreshHint(webRuntime);
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
  if (headline) {
    headline.textContent = 'Loading state...';
  }
  initRefreshControls();
  await reloadState(false);
  applyAutoRefreshTimer();
}

boot();
