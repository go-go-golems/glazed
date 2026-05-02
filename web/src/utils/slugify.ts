import type { ReactNode } from 'react';

export function slugifyHeading(text: string): string {
  const slug = text
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/[\s-]+/g, '-')
    .replace(/^-+|-+$/g, '');
  return slug || 'section';
}

export function textFromReactNode(node: ReactNode): string {
  if (node == null || typeof node === 'boolean') return '';
  if (typeof node === 'string' || typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(textFromReactNode).join('');
  if (typeof node === 'object' && 'props' in node) {
    return textFromReactNode((node as { props?: { children?: ReactNode } }).props?.children);
  }
  return '';
}
