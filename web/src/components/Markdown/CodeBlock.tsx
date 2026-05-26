// CodeBlock.tsx — wraps <pre> elements with a copy-to-clipboard button.
import { useCallback, useRef, useState, type ComponentPropsWithoutRef, type ReactNode } from 'react';

interface CodeBlockProps extends ComponentPropsWithoutRef<'pre'> {
  children?: ReactNode;
}

export function CodeBlock({ children, ...props }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const preRef = useRef<HTMLPreElement>(null);

  const handleCopy = useCallback(async () => {
    // Extract text from the <code> child inside the <pre>, or fall back to
    // the <pre> textContent directly.
    const codeEl = preRef.current?.querySelector('code');
    const text = codeEl?.textContent ?? preRef.current?.textContent ?? '';
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Silently fail — clipboard API may not be available in all contexts
      // (e.g. iframes without focus, older browsers).
    }
  }, []);

  return (
    <div className="code-block-wrapper">
      <pre ref={preRef} {...props}>
        {children}
      </pre>
      <button
        type="button"
        className="copy-button"
        onClick={handleCopy}
        aria-label="Copy code to clipboard"
        title={copied ? 'Copied!' : 'Copy code'}
      >
        {copied ? '✓' : '📋'}
      </button>
    </div>
  );
}
