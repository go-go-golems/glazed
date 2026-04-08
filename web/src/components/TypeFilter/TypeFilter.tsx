// TypeFilter.tsx — row of filter buttons: All, Topic, Example, App, Tutorial.
import { TypeFilterParts } from './parts';
import type { SectionType } from '../../types';
import './styles/typefilter.css';

type FilterValue = 'All' | SectionType;

const FILTERS: FilterValue[] = ['All', 'GeneralTopic', 'Example', 'Application', 'Tutorial'];

/** Maps section types to short button labels. */
const FILTER_LABELS: Record<FilterValue, string> = {
  All:           'All',
  GeneralTopic:  'Topic',
  Example:       'Example',
  Application:   'App',
  Tutorial:      'Tutorial',
};

interface TypeFilterProps {
  value: FilterValue;
  onChange: (value: FilterValue) => void;
}

export function TypeFilter({ value, onChange }: TypeFilterProps) {
  return (
    <div data-part={TypeFilterParts.root} role="group" aria-label="Filter by type">
      {FILTERS.map((f) => (
        <button
          key={f}
          data-part={TypeFilterParts.button}
          aria-pressed={value === f}
          onClick={() => onChange(f)}
        >
          {FILTER_LABELS[f]}
        </button>
      ))}
    </div>
  );
}

export type { FilterValue };
