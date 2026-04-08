// AppLayout.tsx — two-pane layout container.
import type { ReactNode } from 'react';
import { AppLayoutParts } from './parts';
import './styles/app-layout.css';

interface AppLayoutProps {
  sidebar: ReactNode;
  content: ReactNode;
}

export function AppLayout({ sidebar, content }: AppLayoutProps) {
  return (
    <div data-part={AppLayoutParts.root}>
      <div data-part={AppLayoutParts.sidebar}>{sidebar}</div>
      <div data-part={AppLayoutParts.content}>{content}</div>
    </div>
  );
}
