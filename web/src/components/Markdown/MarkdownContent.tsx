// MarkdownContent.tsx — renders Markdown using react-markdown with GFM support.
import type { ComponentPropsWithoutRef, ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { slugifyHeading, textFromReactNode } from '../../utils/slugify';
import { MarkdownContentParts } from './parts';
import './styles/markdown.css';

interface MarkdownContentProps {
  /** Raw Markdown string. */
  content: string;
}

type HeadingTag = 'h1' | 'h2' | 'h3' | 'h4';

type HeadingProps = ComponentPropsWithoutRef<'h1'> & {
  children?: ReactNode;
};

function createHeading(tag: HeadingTag) {
  return function Heading({ children, ...props }: HeadingProps) {
    const Tag = tag;
    const id = slugifyHeading(textFromReactNode(children));
    return <Tag id={id} {...props}>{children}</Tag>;
  };
}

const markdownComponents = {
  h1: createHeading('h1'),
  h2: createHeading('h2'),
  h3: createHeading('h3'),
  h4: createHeading('h4'),
};

export function MarkdownContent({ content }: MarkdownContentProps) {
  return (
    <div data-part={MarkdownContentParts.root}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
        {content}
      </ReactMarkdown>
    </div>
  );
}
