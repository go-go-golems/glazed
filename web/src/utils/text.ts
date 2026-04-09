// utils/text.ts — text utilities for display.

/**
 * Strip common markdown formatting from a string to produce clean plain text.
 * Removes code fences, inline code, bold, italic, and links.
 */
export function stripMarkdown(text: string): string {
  let out = text;

  // Remove fenced code blocks (```...```)
  out = out.replace(/```[\s\S]*?```/g, '');

  // Remove inline code (`...`)
  out = out.replace(/`([^`]+)`/g, '$1');

  // Remove bold (**...**)
  out = out.replace(/\*\*(.+?)\*\*/g, '$1');

  // Remove italic (*...*)
  out = out.replace(/\*(.+?)\*/g, '$1');

  // Remove links [text](url)
  out = out.replace(/\[([^\]]+)\]\([^)]+\)/g, '$1');

  // Collapse whitespace
  out = out.replace(/\s+/g, ' ').trim();

  return out;
}
