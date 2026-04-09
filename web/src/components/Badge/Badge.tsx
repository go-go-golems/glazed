// Badge.tsx — coloured tag for section types, topics, commands, and flags.
import { BadgeParts } from './parts';
import './styles/badge.css';

export type BadgeVariant = 'type' | 'topic' | 'command' | 'flag';

interface BadgeProps {
  text: string;
  /** Controls colour and label mapping. Default: 'topic'. */
  variant?: BadgeVariant;
}

/** Maps section type strings to label and colour. Mirrors TYPE_BADGE in the JSX prototype. */
const TYPE_META: Record<string, { label: string; color: string }> = {
  GeneralTopic: { label: 'Topic', color: '#4a7c59' },
  Example:       { label: 'Example', color: '#b8860b' },
  Application:   { label: 'App',    color: '#4a6a8c' },
  Tutorial:     { label: 'Tutorial', color: '#8c4a6a' },
};

const VARIANT_COLORS: Record<Exclude<BadgeVariant, 'type'>, string> = {
  topic:   '#000',
  command: '#4a7c59',
  flag:    '#8c4a6a',
};

export function Badge({ text, variant = 'topic' }: BadgeProps) {
  const isType = variant === 'type';
  const color = isType
    ? (TYPE_META[text]?.color ?? '#000')
    : (VARIANT_COLORS[variant] ?? '#000');
  const weight = isType ? '700' : '400';
  const displayText = isType ? (TYPE_META[text]?.label ?? text) : text;

  return (
    <span
      data-part={BadgeParts.root}
      data-variant={variant}
      style={{
        '--badge-color': color,
        '--badge-weight': weight,
      } as React.CSSProperties}
    >
      {displayText}
    </span>
  );
}
