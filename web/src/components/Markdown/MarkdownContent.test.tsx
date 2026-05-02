import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { MarkdownContent } from './MarkdownContent';

describe('MarkdownContent', () => {
  it('generates unique IDs for repeated headings', () => {
    const { container } = render(
      <MarkdownContent content={'## Install\n\n### Install\n\n## Install\n'} />,
    );

    const headings = Array.from(container.querySelectorAll('h2, h3')).map((heading) => ({
      id: heading.id,
      text: heading.textContent,
    }));

    expect(headings).toEqual([
      { id: 'install', text: 'Install' },
      { id: 'install-2', text: 'Install' },
      { id: 'install-3', text: 'Install' },
    ]);
  });
});
