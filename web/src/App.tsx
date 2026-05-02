// App.tsx — root component: wires all components together with RTK Query state.
import { useEffect, useMemo, useState } from 'react';
import { matchPath, useLocation, useNavigate } from 'react-router-dom';
import { TitleBar } from './components/TitleBar/TitleBar';
import { SearchBar } from './components/SearchBar/SearchBar';
import { TypeFilter, type FilterValue } from './components/TypeFilter/TypeFilter';
import { PackageSelector } from './components/PackageSelector/PackageSelector';
import { SectionList } from './components/SectionList/SectionList';
import { SectionView } from './components/SectionView/SectionView';
import { EmptyState } from './components/EmptyState/EmptyState';
import { StatusBar } from './components/StatusBar/StatusBar';
import { AppLayout } from './components/AppLayout/AppLayout';
import { useListPackagesQuery, useListSectionsQuery, useGetSectionQuery } from './services/api';
import type { SectionSummary } from './types';

export default function App() {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<FilterValue>('All');
  const [selectedPackage, setSelectedPackage] = useState('');
  const [selectedVersion, setSelectedVersion] = useState('');
  const location = useLocation();
  const navigate = useNavigate();

  const activeSlug = useMemo(() => {
    const match = matchPath('/sections/:slug', location.pathname);
    return match?.params.slug ?? null;
  }, [location.pathname]);

  const { data: packageData } = useListPackagesQuery();
  const packages = packageData?.packages ?? [];
  const currentPackage = packages.find((pkg) => pkg.name === selectedPackage);
  const currentVersions = currentPackage?.versions ?? [];
  const effectiveVersion = currentVersions.length ? selectedVersion : '';

  useEffect(() => {
    if (!packageData || selectedPackage) return;
    const initialPackage = packageData.defaultPackage || packageData.packages[0]?.name || '';
    const initial = packageData.packages.find((pkg) => pkg.name === initialPackage);
    const initialVersions = initial?.versions ?? [];
    setSelectedPackage(initialPackage);
    setSelectedVersion(packageData.defaultVersion || initialVersions[0] || '');
  }, [packageData, selectedPackage]);

  const handlePackageChange = (value: string) => {
    const nextPackage = packages.find((pkg) => pkg.name === value);
    const nextVersions = nextPackage?.versions ?? [];
    setSelectedPackage(value);
    setSelectedVersion(nextVersions[0] || '');
  };

  const { data: listData, isLoading, error } = useListSectionsQuery(
    selectedPackage ? { packageName: selectedPackage, version: effectiveVersion } : undefined,
  );
  const { data: section } = useGetSectionQuery({
    slug: activeSlug!,
    packageName: selectedPackage,
    version: effectiveVersion,
  }, {
    skip: !activeSlug || !selectedPackage,
  });

  const handleSelect = (slug: string) => {
    navigate(`/sections/${slug}`);
  };

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
            <TitleBar title="📁 Sections" />
            <div style={{ padding: '10px 10px 8px', borderBottom: '2px solid #000' }}>
              <div style={{ marginBottom: 8 }}>
                <SearchBar value={search} onChange={setSearch} />
              </div>
              <PackageSelector
                packages={packages}
                selectedPackage={selectedPackage}
                selectedVersion={effectiveVersion}
                onPackageChange={handlePackageChange}
                onVersionChange={setSelectedVersion}
              />
              <TypeFilter value={filter} onChange={setFilter} />
            </div>
            <SectionList
              sections={filtered}
              activeSlug={activeSlug}
              onSelect={handleSelect}
            />
            <StatusBar count={filtered.length} />
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
              <div style={{ flex: 1, overflow: 'auto' }}>
                <SectionView section={section} />
              </div>
            )}
            {!isLoading && !error && !section && (
              <EmptyState />
            )}
          </>
        }
      />
    </div>
  );
}
