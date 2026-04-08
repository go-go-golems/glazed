// MarkdownContent.tsx — renders Markdown using react-markdown with GFM support.
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { MarkdownContentParts } from './parts';
import './styles/markdown.css';

interface MarkdownContentProps {
  /** Raw Markdown string. */
  content: string;
}

export function MarkdownContent({ content }: MarkdownContentProps) {
  return (
    <div data-part={MarkdownContentParts.root}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>
        {content}
      </ReactMarkdown>
    </div>
  );
}
