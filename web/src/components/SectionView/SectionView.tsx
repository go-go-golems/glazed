// SectionView.tsx — full section rendering: header + Markdown body.
import type { SectionDetail } from '../../types';
import { SectionHeader } from './SectionHeader';
import { MarkdownContent } from '../Markdown/MarkdownContent';
import { SectionViewParts } from './parts';
import './styles/section-view.css';

interface SectionViewProps {
  section: SectionDetail;
}

export function SectionView({ section }: SectionViewProps) {
  return (
    <div data-part={SectionViewParts.root}>
      <SectionHeader section={section} />
      <div data-part={SectionViewParts.body}>
        <MarkdownContent content={section.content} />
      </div>
    </div>
  );
}
