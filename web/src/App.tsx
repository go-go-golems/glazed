// App.tsx — root component: wires all components together with RTK Query state.
//
// URL scheme: /{package}/{version}/sections/{slug}#{heading-id}
//
// Package and version are always in the URL path. For unversioned packages,
// the version segment is "_" (a placeholder meaning "no version").
// The slug is optional; when absent, the section list is shown with EmptyState.
// The heading fragment (#heading-id) scrolls to a subsection after the section loads.

import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { TitleBar } from './components/TitleBar/TitleBar';
import { SearchBar } from './components/SearchBar/SearchBar';
import { TypeFilter, type FilterValue } from './components/TypeFilter/TypeFilter';
import { PackageSelector } from './components/PackageSelector/PackageSelector';
import { NavigationModeToggle, type NavigationMode } from './components/NavigationModeToggle/NavigationModeToggle';
import { DocumentationTree } from './components/DocumentationTree/DocumentationTree';
import { SectionList } from './components/SectionList/SectionList';
import { SectionView } from './components/SectionView/SectionView';
import { EmptyState } from './components/EmptyState/EmptyState';
import { PackageIndex } from './components/PackageIndex/PackageIndex';
import { StatusBar } from './components/StatusBar/StatusBar';
import { AppLayout } from './components/AppLayout/AppLayout';
import { useListPackagesQuery, useListSectionsQuery, useGetSectionQuery } from './services/api';
import type { SectionSummary } from './types';

/** URL segment used for packages that have no version. */
const NO_VERSION = '_';

/**
 * Convert a version string from the URL to the API representation.
 * The URL uses "_" as a placeholder for "no version", which the API
 * expects as an empty string.
 */
function versionFromUrl(urlVersion: string | undefined): string {
  if (!urlVersion || urlVersion === NO_VERSION) return '';
  return urlVersion;
}

/**
 * Convert a version string from the API to the URL representation.
 * Empty versions become "_" so the URL always has two segments after the root.
 */
function versionToUrl(apiVersion: string): string {
  return apiVersion || NO_VERSION;
}

export default function App() {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<FilterValue>('All');
  const [navigationMode, setNavigationMode] = useState<NavigationMode>('tree');
  const location = useLocation();
  const navigate = useNavigate();
  const contentScrollRef = useRef<HTMLDivElement>(null);

  // Extract package, version, and slug from URL params.
  const params = useParams();
  const urlPackage = params['package'] ?? '';
  const urlVersion = params.version ?? '';
  const activeSlug = params.slug ?? null;
  const activeHeadingId = location.hash ? location.hash.replace(/^#/, '') : '';

  const selectedPackage = urlPackage;
  const selectedVersion = versionFromUrl(urlVersion);
  const effectiveVersion = selectedVersion;

  const { data: packageData } = useListPackagesQuery();
  const packages = packageData?.packages ?? [];

  // When landing on / (no package in URL), redirect to the default package/version.
  useEffect(() => {
    if (!packageData || selectedPackage) return;
    const defaultPkg = packageData.defaultPackage || packageData.packages[0]?.name || 'default';
    const defaultPkgInfo = packageData.packages.find((pkg) => pkg.name === defaultPkg);
    const defaultVersions = defaultPkgInfo?.versions ?? [];
    const defaultVer = defaultVersions[0] || '';
    navigate(`/${defaultPkg}/${versionToUrl(defaultVer)}`, { replace: true });
  }, [packageData, selectedPackage, navigate]);

  // When landing on a package that doesn't exist, redirect to default.
  useEffect(() => {
    if (!packageData || !selectedPackage) return;
    const exists = packages.some((pkg) => pkg.name === selectedPackage);
    if (!exists && packages.length > 0) {
      const defaultPkg = packageData.defaultPackage || packages[0].name;
      const defaultPkgInfo = packages.find((pkg) => pkg.name === defaultPkg);
      const defaultVer = defaultPkgInfo?.versions?.[0] || '';
      navigate(`/${defaultPkg}/${versionToUrl(defaultVer)}`, { replace: true });
    }
  }, [packageData, selectedPackage, packages, navigate]);

  const handlePackageChange = (value: string) => {
    const nextPackage = packages.find((pkg) => pkg.name === value);
    // Prefer explicit latestVersion from the API; fall back to versions[0].
    // The server reverse-sorts versions so versions[0] is the latest, but
    // latestVersion makes the contract explicit.
    const newVersion = nextPackage?.latestVersion
      || nextPackage?.versions?.[0]
      || '';
    navigate(`/${value}/${versionToUrl(newVersion)}`);
  };

  const handleVersionChange = (value: string) => {
    navigate(`/${selectedPackage}/${versionToUrl(value)}`);
  };

  const { data: listData, isLoading, error } = useListSectionsQuery(
    selectedPackage ? { packageName: selectedPackage, version: effectiveVersion } : undefined,
  );
  const { data: sectionData } = useGetSectionQuery({
    slug: activeSlug!,
    packageName: selectedPackage,
    version: effectiveVersion,
  }, {
    skip: !activeSlug || !selectedPackage,
  });
  // Only use section data when a slug is actually active in the URL.
  // When skip is true, useGetSectionQuery returns stale cached data;
  // we must null it out when no slug is selected.
  const section = activeSlug ? sectionData : undefined;

  const handleSelect = (slug: string) => {
    navigate(`/${selectedPackage}/${versionToUrl(effectiveVersion)}/sections/${slug}`);
  };

  const handleSelectHeading = (slug: string, headingId: string) => {
    navigate(`/${selectedPackage}/${versionToUrl(effectiveVersion)}/sections/${slug}#${headingId}`);
  };

  useEffect(() => {
    if (!section) return;

    requestAnimationFrame(() => {
      if (activeHeadingId) {
        const heading = document.getElementById(activeHeadingId);
        if (heading) {
          heading.scrollIntoView({ block: 'start' });
          return;
        }
      }

      if (contentScrollRef.current) {
        contentScrollRef.current.scrollTop = 0;
      }
    });
  }, [section, activeHeadingId]);

  // Client-side filter — mirrors the JSX prototype logic.
  const filtered = useMemo(() => {
    if (!listData) return [];
    return (listData.sections ?? []).filter((s: SectionSummary) => {
      if (filter !== 'All' && s.type !== filter) return false;
      if (!search) return true;
      const q = search.toLowerCase();
      return (
        s.title.toLowerCase().includes(q) ||
        s.short.toLowerCase().includes(q) ||
        s.topics.some((t: string) => t.toLowerCase().includes(q)) ||
        s.slug.toLowerCase().includes(q)
      );
    });
  }, [listData, search, filter]);

  return (
    <div className="app-root">
      <AppLayout
        sidebar={
          <>
            <TitleBar title="📖 Documentation" />
            <div style={{ padding: '10px 10px 8px', borderBottom: '2px solid #000' }}>
              <PackageSelector
                packages={packages}
                selectedPackage={selectedPackage}
                selectedVersion={effectiveVersion}
                onPackageChange={handlePackageChange}
                onVersionChange={handleVersionChange}
              />
              <div style={{ marginBottom: 8 }}>
                <SearchBar value={search} onChange={setSearch} placeholder="Search documentation…" />
              </div>
              <NavigationModeToggle value={navigationMode} onChange={setNavigationMode} />
              {navigationMode === 'search' && (
                <TypeFilter value={filter} onChange={setFilter} />
              )}
            </div>
            {navigationMode === 'tree' ? (
              <DocumentationTree
                sections={listData?.sections ?? []}
                search={search}
                activeSlug={activeSlug}
                activeHeadingId={activeHeadingId}
                onSelectDocument={handleSelect}
                onSelectHeading={handleSelectHeading}
              />
            ) : (
              <SectionList
                sections={filtered}
                activeSlug={activeSlug}
                onSelect={handleSelect}
              />
            )}
            <StatusBar count={navigationMode === 'tree' ? (listData?.sections.length ?? 0) : filtered.length} />
          </>
        }
        content={
          <>
            <TitleBar
              title={
                section
                  ? `📄 ${section.title} — glaze help ${section.slug}`
                  : '📄 Documentation'
              }
            />
            {isLoading && (
              <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                Loading…
              </div>
            )}
            {error && (
              <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'red' }}>
                Error loading sections.
              </div>
            )}
            {!isLoading && !error && section && (
              <div ref={contentScrollRef} style={{ flex: 1, overflow: 'auto' }}>
                <SectionView section={section} />
              </div>
            )}
            {!isLoading && !error && !section && listData && listData.sections.length > 0 && (
              <div ref={contentScrollRef} style={{ flex: 1, overflow: 'auto' }}>
                <PackageIndex
                  sections={listData.sections}
                  onSelect={handleSelect}
                />
              </div>
            )}
            {!isLoading && !error && !section && (!listData || listData.sections.length === 0) && (
              <EmptyState />
            )}
          </>
        }
      />
    </div>
  );
}
