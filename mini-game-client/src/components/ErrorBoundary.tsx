import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

/**
 * Error Boundary Component
 * Catches JavaScript errors anywhere in the child component tree
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div style={styles.container}>
          <div style={styles.content}>
            <h1 style={styles.title}>⚠️ 오류가 발생했습니다</h1>
            <p style={styles.message}>
              {this.state.error?.message || '알 수 없는 오류가 발생했습니다.'}
            </p>
            <button
              style={styles.button}
              onClick={() => window.location.href = '/'}
            >
              홈으로 돌아가기
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

const styles = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
    padding: '20px',
  },
  content: {
    textAlign: 'center' as const,
    backgroundColor: 'white',
    padding: '40px',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
    maxWidth: '500px',
  },
  title: {
    fontSize: '32px',
    fontWeight: 'bold',
    marginBottom: '20px',
    color: '#e74c3c',
  },
  message: {
    fontSize: '16px',
    color: '#666',
    marginBottom: '30px',
    lineHeight: '1.5',
  },
  button: {
    backgroundColor: '#3498db',
    color: 'white',
    border: 'none',
    padding: '12px 30px',
    fontSize: '16px',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold',
  },
};
