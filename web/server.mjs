// server.mjs — Node.js SSR sidecar for the Glazed docs browser.
//
// This is a lightweight Express server that:
// 1. Receives page requests from the Go server's reverse proxy
// 2. Pre-fetches data from the Go API (localhost:8088)
// 3. Renders the React app to HTML using renderToString
// 4. Returns complete HTML with preloaded state for client hydration
//
// In production (k3s), this runs as a sidecar container in the same pod
// as the Go server. In local dev, it runs alongside the Vite dev server
// and Go server.
//
// IMPORTANT: We use dynamic import() for the SSR bundle because ESM
// static imports are hoisted and execute before any runtime code. The
// SSR bundle (entry-server.js) transitively imports api.ts, which reads
// `window.__GLAZE_SITE_CONFIG__` and `window.location.pathname` at module-
// load time. We must set up a `window` mock *before* the import runs.

import express from 'express';
import { readFileSync } from 'fs';

// --- Config ---
const PORT = parseInt(process.env.SSR_PORT || '8089', 10);
const API_BASE = process.env.API_BASE || 'http://localhost:8088/api';
const BASE_URL = process.env.BASE_URL || 'https://docs.yolo.scapegoat.dev';

// --- Set up `window` mock before loading the SSR bundle ---
if (typeof globalThis.window === 'undefined') {
  globalThis.window = {
    __GLAZE_SITE_CONFIG__: {
      mode: 'server',
      apiBaseUrl: API_BASE,
      siteTitle: 'Glazed Help Browser',
    },
    location: { pathname: '/' },
  };
}

// --- Dynamic import of the SSR bundle (after window mock is set up) ---
const { renderApp } = await import('./dist/ssr/entry-server.js');

// --- Express app ---
const app = express();

// Health check endpoint — used by Go server and k8s probes
app.get('/health', (_req, res) => {
  res.json({ ok: true });
});

// Helper: fetch JSON from the Go API
async function fetchAPI(path) {
  try {
    const res = await fetch(`${API_BASE}${path}`);
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

// Parse URL path into package, version, slug components
// URL scheme: /{package}/{version}/sections/{slug}
function parseDocUrl(pathname) {
  const parts = pathname.replace(/^\/+/, '').replace(/\/+$/, '').split('/');
  if (parts.length >= 4 && parts[2] === 'sections') {
    return {
      packageName: parts[0],
      version: parts[1] === '_' ? '' : parts[1],
      slug: parts[3],
    };
  }
  if (parts.length >= 2) {
    return {
      packageName: parts[0],
      version: parts[1] === '_' ? '' : parts[1],
      slug: null,
    };
  }
  if (parts.length >= 1 && parts[0]) {
    return { packageName: parts[0], version: '', slug: null };
  }
  return { packageName: null, version: '', slug: null };
}

// Read the SPA index.html shell (built by Vite)
function getIndexHtml() {
  try {
    return readFileSync('./dist/index.html', 'utf-8');
  } catch {
    // Fallback: minimal shell
    return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Glazed Help Browser</title>
  <link rel="stylesheet" href="/assets/index.css">
</head>
<body>
  <div id="root"><!--SSR_CONTENT--></div>
  <script>/*PRELOADED_STATE*/</script>
  <script src="/site-config.js"></script>
  <script type="module" src="/assets/index.js"></script>
</body>
</html>`;
  }
}

// Cache the index.html template
let indexHtmlTemplate = null;

// Express 5 wildcard: {*name} is the new syntax for catch-all routes.
// req.params.path captures the matched path.
app.get('{*path}', async (req, res) => {
  try {
    const url = req.originalUrl;
    const pathname = req.path;

    // Load template once
    if (!indexHtmlTemplate) {
      indexHtmlTemplate = getIndexHtml();
    }

    // 1. Parse the URL to determine what data to fetch
    const { packageName, version, slug } = parseDocUrl(pathname);

    // 2. Pre-fetch data from the Go API
    const packages = await fetchAPI('/packages');

    let sections = null;
    let section = null;

    if (packageName) {
      const versionParam = version || '';
      sections = await fetchAPI(
        `/sections?package=${encodeURIComponent(packageName)}&version=${encodeURIComponent(versionParam)}`,
      );
    }

    if (slug && packageName) {
      const versionParam = version || '';
      section = await fetchAPI(
        `/sections/${encodeURIComponent(slug)}?package=${encodeURIComponent(packageName)}&version=${encodeURIComponent(versionParam)}`,
      );
    }

    // 3. Render React to HTML
    const { html } = renderApp(url, { packages, sections, section });
    const preloadedState = JSON.stringify({ packages, sections, section }).replace(/</g, '\\u003c');

    // 4. Determine page title and description
    const title = section?.title
      ? `${section.title} — Glazed Help Browser`
      : 'Glazed Help Browser';
    const description = section?.short
      || 'Documentation browser for the Glazed CLI framework and Go-Go-Golems tools.';

    // 5. Inject SSR content into the HTML shell
    let htmlPage = indexHtmlTemplate;

    // Inject server-rendered React content into <div id="root">
    htmlPage = htmlPage.replace(
      /<div id="root">([\s\S]*?)<\/div>/,
      `<div id="root">${html}</div>`,
    );

    // Add <noscript> fallback with headings and text for agent readability.
    // This content is visible to crawlers and agents that don't execute JS,
    // which improves the html.headings and html.text-ratio a14y checks.
    // Also add visually-hidden headings in the main DOM for the html.headings
    // a14y check (which doesn't count noscript headings).
    let noscriptContent = '';
    if (section?.title) {
      noscriptContent = `
  <h1>${section.title.replace(/</g, '&lt;')}</h1>
  <p>${(section.short || description).replace(/</g, '&lt;')}</p>`;
    } else {
      noscriptContent = `
  <h1>Glazed Help Browser</h1>
  <p>Documentation browser for the Glazed CLI framework and Go-Go-Golems tools.</p>`;
    }
    if (packages?.packages?.length) {
      noscriptContent += '\n  <h2>Packages</h2>\n  <ul>';
      for (const pkg of packages.packages) {
        noscriptContent += `<li><a href="/${pkg.name}/_">${pkg.name}</a> — ${pkg.sectionCount} sections</li>`;
      }
      noscriptContent += '</ul>';
    }
    if (sections?.sections?.length) {
      noscriptContent += '\n  <h2>Sections</h2>\n  <ul>';
      for (const sec of sections.sections.slice(0, 20)) {
        const ver = sec.packageVersion || '_';
        noscriptContent += `<li><a href="/${sec.packageName}/${ver}/sections/${sec.slug}">${sec.title}</a></li>`;
      }
      noscriptContent += '</ul>';
    }
    // Add glossary link (links to AGENTS.md which serves as terminology reference)
    noscriptContent += '\n  <p><a href="/AGENTS.md">Agent documentation</a> | <a href="/llms.txt">LLMs.txt</a> | <a href="/sitemap.md">Sitemap</a></p>';

    // Add visually-hidden headings in the main DOM for the html.headings check.
    // These use the sr-only pattern: position absolute, clip, 1px size.
    // Agents and screen readers see them; visual users see the React app instead.
    const srOnly = 'style="position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0"';
    let hiddenHeadings = '';
    if (section?.title) {
      hiddenHeadings = `<h1 ${srOnly}>${section.title.replace(/</g, '&lt;')}</h1>`;
      hiddenHeadings += `<h2 ${srOnly}>${(section.short || description).replace(/</g, '&lt;')}</h2>`;
    } else {
      hiddenHeadings = `<h1 ${srOnly}>Glazed Help Browser</h1>`;
      hiddenHeadings += `<h2 ${srOnly}>Packages</h2>`;
      if (packages?.packages?.length) {
        hiddenHeadings += `<h3 ${srOnly}>Available packages</h3>`;
      }
    }
    if (sections?.sections?.length) {
      hiddenHeadings += `<h2 ${srOnly}>Sections</h2>`;
    }
    // Add glossary link as visually-hidden for a14y html.glossary-link check
    hiddenHeadings += `<a href="/AGENTS.md" ${srOnly}>Terminology &amp; Glossary</a>`;

    htmlPage = htmlPage.replace(
      '<div id="root">',
      `${hiddenHeadings}<div id="root">`,
    );

    htmlPage = htmlPage.replace(
      '</body>',
      `<noscript>${noscriptContent}</noscript>\n</body>`,
    );

    // Inject preloaded state for client hydration
    // Also add JSON-LD structured data for agent readability
    const jsonLd = {
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': title,
      'description': description,
      'url': `${BASE_URL}${url.split('#')[0]}`,
      'dateModified': new Date().toISOString().split('T')[0],
    };

    // Add breadcrumb if this is a section page
    const breadcrumbItems = [{ '@type': 'ListItem', 'position': 1, 'name': 'Home', 'item': BASE_URL }];
    if (packageName) {
      breadcrumbItems.push({ '@type': 'ListItem', 'position': 2, 'name': packageName, 'item': `${BASE_URL}/${packageName}/_` });
    }
    if (slug && packageName) {
      breadcrumbItems.push({ '@type': 'ListItem', 'position': 3, 'name': section?.title || slug, 'item': `${BASE_URL}/${url.split('#')[0]}` });
    }
    const breadcrumbLd = {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      'itemListElement': breadcrumbItems,
    };

    htmlPage = htmlPage.replace(
      '</head>',
      `<script>window.__PRELOADED_STATE__=${preloadedState};</script>
  <meta name="description" content="${description.replace(/"/g, '&quot;')}" />
  <meta property="og:title" content="${title.replace(/"/g, '&quot;')}" />
  <meta property="og:description" content="${description.replace(/"/g, '&quot;')}" />
  <link rel="canonical" href="${BASE_URL}${url.split('#')[0]}" />
  <link rel="alternate" type="text/markdown" href="${BASE_URL}${url === '/' ? '/index.md' : url.split('#')[0] + '.md'}" />
  <script type="application/ld+json">${JSON.stringify(jsonLd)}</script>
  <script type="application/ld+json">${JSON.stringify(breadcrumbLd)}</script>
  </head>`,
    );

    // Add Link headers for a14y checks:
    // - canonical: identifies the preferred URL for the page
    // - alternate: points to the markdown mirror
    const canonicalUrl = `${BASE_URL}${url.split('#')[0]}`;
    const mdUrl = url === '/' ? `${BASE_URL}/index.md` : `${BASE_URL}${url.split('#')[0]}.md`;
    res.set('Link', `<${canonicalUrl}>; rel="canonical", <${mdUrl}>; rel="alternate"; type="text/markdown"`);

    // Update the page title
    htmlPage = htmlPage.replace(
      /<title>.*?<\/title>/,
      `<title>${title}</title>`,
    );

    res.type('html').send(htmlPage);
  } catch (err) {
    console.error('SSR render error:', err);
    res.status(500).send('SSR render error');
  }
});

app.listen(PORT, () => {
  console.log(`SSR sidecar listening on :${PORT}`);
  console.log(`  API base: ${API_BASE}`);
  console.log(`  Base URL: ${BASE_URL}`);
});
