// components/TypographyPalette/css-override-engine.ts
// Generates and injects CSS rules for typography overrides.
// Also provides export functions for clipboard copy.

import type { TypographyOverrides, TypographyProperties } from '../../types/typography-palette';
import { FONT_STACKS } from '../../types/typography-palette';
import { buildElementMap } from './element-registry';

const STYLE_ID = 'typography-palette-overrides';

/** Apply all overrides to the DOM by injecting a <style> element. */
export function applyOverrides(overrides: TypographyOverrides): void {
  const elementMap = buildElementMap();
  const rules: string[] = [];

  for (const elementId of Object.keys(overrides)) {
    const rule = generateRule(elementId, overrides, elementMap);
    if (rule) rules.push(rule);
  }

  let styleEl = document.getElementById(STYLE_ID) as HTMLStyleElement | null;
  if (!styleEl) {
    styleEl = document.createElement('style');
    styleEl.id = STYLE_ID;
    document.head.appendChild(styleEl);
  }

  styleEl.textContent = rules.length > 0 ? rules.join('\n\n') : '';
}

/** Remove all overrides from the DOM. */
export function clearOverrides(): void {
  const styleEl = document.getElementById(STYLE_ID);
  if (styleEl) {
    styleEl.textContent = '';
  }
}

/** Generate a single CSS rule string for an element override. */
function generateRule(
  elementId: string,
  overrides: TypographyOverrides,
  elementMap: Map<string, { selector: string }>,
): string | null {
  const elem = elementMap.get(elementId);
  if (!elem) return null;

  const props = overrides[elementId];
  if (!props) return null;

  const declarations = buildDeclarations(props);
  if (declarations.length === 0) return null;

  return `${elem.selector} {\n${declarations.join('\n')}\n}`;
}

/** Build CSS declaration strings from TypographyProperties. */
function buildDeclarations(props: TypographyProperties): string[] {
  const declarations: string[] = [];

  if (props.fontFamily !== undefined) {
    const stack = FONT_STACKS[props.fontFamily] || FONT_STACKS['ui'];
    declarations.push(`  font-family: ${stack};`);
  }
  if (props.fontSize !== undefined) {
    const unit = props.fontSizeUnit || 'px';
    declarations.push(`  font-size: ${props.fontSize}${unit};`);
  }
  if (props.fontWeight !== undefined) {
    declarations.push(`  font-weight: ${props.fontWeight};`);
  }
  if (props.color !== undefined) {
    declarations.push(`  color: ${props.color};`);
  }
  if (props.lineHeight !== undefined) {
    declarations.push(`  line-height: ${props.lineHeight};`);
  }

  return declarations;
}

// ---------------------------------------------------------------------------
// Export to clipboard
// ---------------------------------------------------------------------------

/** Generate the full CSS text for the current overrides (for clipboard export). */
export function generateCssExport(overrides: TypographyOverrides): string {
  const elementMap = buildElementMap();
  const rules: string[] = [];

  for (const elementId of Object.keys(overrides)) {
    const rule = generateRule(elementId, overrides, elementMap);
    if (rule) rules.push(rule);
  }

  if (rules.length === 0) return '/* No typography overrides active */';

  return [
    '/* Typography Debug Palette — exported overrides */',
    '/* Paste these rules into your component CSS files or */',
    '/* use as a guide to update CSS custom properties.    */',
    '',
    ...rules,
  ].join('\n');
}

/** Generate a CSS custom properties block from overrides.
 *  This format is useful for updating global.css :root variables. */
export function generateCssVariablesExport(overrides: TypographyOverrides): string {
  const lines: string[] = [];

  // Map element IDs to logical CSS variable names
  const varMap: Record<string, Record<string, string>> = {
    'root.body':          { fontSize: '--font-size-base', fontFamily: '--font-ui', color: '--color-fg' },
    'prose.body':         { lineHeight: '--prose-line-height' },
    'code.inline':        { fontFamily: '--font-mono' },
    'code.block':         { fontFamily: '--font-mono' },
    'extras.link':        { color: '--color-accent' },
  };

  for (const [elementId, props] of Object.entries(overrides)) {
    const mappings = varMap[elementId];
    if (!mappings) continue;

    for (const [prop, varName] of Object.entries(mappings)) {
      const value = getPropertyValue(props, prop);
      if (value !== undefined) {
        lines.push(`  ${varName}: ${value};`);
      }
    }
  }

  if (lines.length === 0) {
    // Fall back to the per-rule export
    return generateCssExport(overrides);
  }

  return [
    '/* Typography Debug Palette — CSS variable overrides */',
    '/* Apply to :root { ... } in global.css */',
    '',
    ':root {',
    ...lines,
    '}',
  ].join('\n');
}

/** Get a formatted property value from TypographyProperties by key name. */
function getPropertyValue(props: TypographyProperties, prop: string): string | undefined {
  switch (prop) {
    case 'fontFamily':
      return props.fontFamily ? FONT_STACKS[props.fontFamily] : undefined;
    case 'fontSize':
      return props.fontSize !== undefined ? `${props.fontSize}${props.fontSizeUnit || 'px'}` : undefined;
    case 'fontWeight':
      return props.fontWeight !== undefined ? String(props.fontWeight) : undefined;
    case 'color':
      return props.color;
    case 'lineHeight':
      return props.lineHeight !== undefined ? String(props.lineHeight) : undefined;
    default:
      return undefined;
  }
}

/** Copy CSS export text to the clipboard. Returns true on success. */
export async function copyCssToClipboard(overrides: TypographyOverrides, format: 'rules' | 'variables' = 'rules'): Promise<boolean> {
  const text = format === 'variables'
    ? generateCssVariablesExport(overrides)
    : generateCssExport(overrides);

  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
