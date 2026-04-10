// App.tsx — root component: wires all components together with RTK Query state.
import { useState, useMemo } from 'react';
import { matchPath, useLocation, useNavigate } from 'react-router-dom';
import { TitleBar } from './components/TitleBar/TitleBar';
import { SearchBar } from './components/SearchBar/SearchBar';
import { TypeFilter, type FilterValue } from './components/TypeFilter/TypeFilter';
import { SectionList } from './components/SectionList/SectionList';
import { SectionView } from './components/SectionView/SectionView';
import { EmptyState } from './components/EmptyState/EmptyState';
import { StatusBar } from './components/StatusBar/StatusBar';
import { AppLayout } from './components/AppLayout/AppLayout';
import { useListSectionsQuery, useGetSectionQuery } from './services/api';
import type { SectionSummary } from './types';

export default function App() {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<FilterValue>('All');
  const location = useLocation();
  const navigate = useNavigate();

  const activeSlug = useMemo(() => {
    const match = matchPath('/sections/:slug', location.pathname);
    return match?.params.slug ?? null;
  }, [location.pathname]);

  const { data: listData, isLoading, error } = useListSectionsQuery();
  const { data: section } = useGetSectionQuery(activeSlug!, {
    skip: !activeSlug,
  });

  const handleSelect = (slug: string) => {
    navigate(`/sections/${slug}`);
  };

  // Client-side filter — mirrors the JSX prototype logic.
  const filtered = useMemo(() => {
    if (!listData) return [];
    return listData.sections.filter((s: SectionSummary) => {
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
