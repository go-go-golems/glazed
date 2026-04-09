// TitleBar.tsx — retro title bar with a square icon and centred title.
import { TitleBarParts } from './parts';
import './styles/titlebar.css';

interface TitleBarProps {
  /** Displayed text, centred in the title bar. */
  title: string;
  /** Icon emoji or string shown in the left box. Default: blank square. */
  icon?: React.ReactNode;
}

export function TitleBar({ title, icon }: TitleBarProps) {
  return (
    <div data-part={TitleBarParts.root}>
      {/* Icon box */}
      <div data-part={TitleBarParts.icon}>
        <div data-part="titlebar-icon-box">
          {icon ?? null}
        </div>
      </div>

      {/* Stripe + title + stripe */}
      <div data-part={TitleBarParts.ruler}>
        <div data-part="titlebar-stripe" />
        <span data-part={TitleBarParts.title}>{title}</span>
        <div data-part="titlebar-stripe" />
      </div>
    </div>
  );
}
