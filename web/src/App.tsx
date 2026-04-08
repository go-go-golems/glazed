// App.tsx
// Root component — placeholder for the full help browser UI.
// The full implementation wires all 13 components together (Tasks 16–28).

import './services/api'; // side-effect: registers RTK Query endpoints
import { useListSectionsQuery } from './services/api';

export default function App() {
  const { data, isLoading, error } = useListSectionsQuery();

  return (
    <div className="app-root">
      <header className="app-header">
        <h1>Glazed Help Browser</h1>
      </header>

      <main>
        {isLoading && <p>Loading…</p>}
        {error && <p style={{ color: 'red' }}>Error loading sections.</p>}
        {data && (
          <p>
            {data.total} section{data.total !== 1 ? 's' : ''} available.
          </p>
        )}
      </main>
    </div>
  );
}
