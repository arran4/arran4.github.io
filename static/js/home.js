(() => {
  'use strict';
  const root = document.documentElement;
  const cards = [...document.querySelectorAll('[data-catalogue-item]')];
  const filters = [...document.querySelectorAll('[data-filter]')];
  const search = document.getElementById('project-search');
  const counter = document.getElementById('project-count');
  const more = document.getElementById('directory-more');
  let category = 'all';
  let expanded = false;

  function refreshCatalogue() {
    if (!search || !counter) return;
    const term = search.value.trim().toLocaleLowerCase();
    const matches = cards.filter(card =>
      (category === 'all' || card.dataset.category === category) &&
      (card.dataset.search || '').toLocaleLowerCase().includes(term)
    );
    const limit = (!expanded && category === 'all' && !term) ? 12 : Infinity;
    cards.forEach(card => {
      const index = matches.indexOf(card);
      card.hidden = index < 0 || index >= limit;
    });
    counter.textContent = matches.length
      ? 'Showing ' + Math.min(matches.length, limit) + ' of ' + matches.length + ' matching projects'
      : 'No projects match this search. Try another term or category.';
    if (more) {
      more.hidden = !!term || category !== 'all' || matches.length <= 12;
      more.textContent = expanded ? 'Show fewer projects' : 'Show all projects';
      more.setAttribute('aria-expanded', String(expanded));
    }
  }
  filters.forEach(button => button.addEventListener('click', () => {
    category = button.dataset.filter;
    expanded = false;
    filters.forEach(item => item.setAttribute('aria-pressed', String(item === button)));
    refreshCatalogue();
  }));
  search?.addEventListener('input', refreshCatalogue);
  more?.addEventListener('click', () => { expanded = !expanded; refreshCatalogue(); });
  refreshCatalogue();

  const releaseCards = [...document.querySelectorAll('[data-release]')];
  const releaseMore = document.getElementById('release-more');
  let showReleases = false;
  function refreshReleases() {
    releaseCards.forEach((card, index) => { card.hidden = !showReleases && index >= 6; });
    if (releaseMore) {
      releaseMore.hidden = releaseCards.length <= 6;
      releaseMore.textContent = showReleases ? 'Show fewer releases' : 'Show all tracked releases';
      releaseMore.setAttribute('aria-expanded', String(showReleases));
    }
  }
  releaseMore?.addEventListener('click', () => { showReleases = !showReleases; refreshReleases(); });
  refreshReleases();

  const modes = ['system', 'dark', 'light'];
  const themeControls = [...document.querySelectorAll('[data-theme-mode]')];
  function setTheme(mode) {
    if (!modes.includes(mode)) mode = 'system';
    if (mode === 'system') root.removeAttribute('data-theme');
    else root.setAttribute('data-theme', mode);
    themeControls.forEach(button => button.setAttribute('aria-pressed', String(button.dataset.themeMode === mode)));
    try { localStorage.setItem('a4-home-theme', mode); } catch (_) { /* storage optional */ }
  }
  let storedTheme = 'system';
  try { storedTheme = localStorage.getItem('a4-home-theme') || 'system'; } catch (_) { /* storage optional */ }
  setTheme(storedTheme);
  themeControls.forEach(button => button.addEventListener('click', () => setTheme(button.dataset.themeMode)));

  const params = new URLSearchParams(location.search);
  if (params.get('preview-support') === '1') {
    document.querySelectorAll('[data-support-preview]').forEach(el => { el.hidden = false; });
  }
})();