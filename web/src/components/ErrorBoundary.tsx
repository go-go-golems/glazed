import React from 'react';

type ErrorBoundaryProps = {
  children: React.ReactNode;
};

type ErrorBoundaryState = {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
};

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = {
    error: null,
    errorInfo: null,
  };

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ error, errorInfo });
    console.error('Uncaught application error', error, errorInfo);
  }

  render() {
    const { error, errorInfo } = this.state;

    if (!error) {
      return this.props.children;
    }

    return (
      <div className="app-root" style={{ padding: 24, fontFamily: 'system-ui, sans-serif' }}>
        <div
          role="alert"
          style={{
            maxWidth: 900,
            border: '2px solid #000',
            background: '#fff',
            boxShadow: '6px 6px 0 #000',
            padding: 20,
          }}
        >
          <h1 style={{ marginTop: 0 }}>Glazed help browser crashed</h1>
          <p>
            The page hit an unexpected client-side error. The server may still be reachable;
            try the health endpoint or reload after the deployment is fixed.
          </p>
          <p>
            <a href="/api/health">Open /api/health</a>
          </p>
          <details open>
            <summary>Error</summary>
            <pre style={{ whiteSpace: 'pre-wrap', overflowX: 'auto' }}>{error.message}</pre>
          </details>
          {errorInfo?.componentStack && (
            <details>
              <summary>Component stack</summary>
              <pre style={{ whiteSpace: 'pre-wrap', overflowX: 'auto' }}>{errorInfo.componentStack}</pre>
            </details>
          )}
        </div>
      </div>
    );
  }
}
