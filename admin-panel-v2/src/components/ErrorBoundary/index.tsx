import { Button, Card, Result } from 'antd';
import React from 'react';
import { withTranslation, type WithTranslation } from 'react-i18next';

function isChunkLoadError(error: Error): boolean {
  return (
    error.name === 'ChunkLoadError' ||
    /(?:loading|failed to load) (?:css )?chunk/i.test(error.message) ||
    /Failed to fetch dynamically imported module/i.test(error.message)
  );
}

function getSubTitleId(isChunkError: boolean, isOffline: boolean): string {
  if (!isChunkError) return 'app.error.render.description';
  return isOffline
    ? 'app.error.chunk.description.offline'
    : 'app.error.chunk.description.online';
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  t: WithTranslation['t'];
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  isOnline: boolean;
  retryCount: number;
}

class ErrorBoundary extends React.Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = {
    hasError: false,
    error: null,
    isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
    retryCount: 0,
  };

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  componentDidMount() {
    window.addEventListener('online', this.handleOnline);
    window.addEventListener('offline', this.handleOffline);
  }

  componentWillUnmount() {
    window.removeEventListener('online', this.handleOnline);
    window.removeEventListener('offline', this.handleOffline);
  }

  handleOnline = () => {
    this.setState({ isOnline: true });
    if (
      this.state.hasError &&
      this.state.error &&
      isChunkLoadError(this.state.error)
    ) {
      window.location.reload();
    }
  };

  handleOffline = () => {
    this.setState({ isOnline: false });
  };

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack);
  }

  handleRetry = () => {
    this.setState((prev) => ({
      hasError: false,
      error: null,
      retryCount: prev.retryCount + 1,
    }));
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (!this.state.hasError || !this.state.error) {
      return (
        <React.Fragment key={this.state.retryCount}>
          {this.props.children}
        </React.Fragment>
      );
    }

    const { t } = this.props;
    const isOffline = !this.state.isOnline;
    const isChunkError = isChunkLoadError(this.state.error);

    return (
      <Card variant="borderless" style={{ margin: 24 }}>
        <Result
          status="error"
          title={t(isChunkError ? 'app.error.chunk.title' : 'app.error.render.title')}
          subTitle={t(getSubTitleId(isChunkError, isOffline))}
          extra={[
            isChunkError && (
              <Button type="primary" key="retry" onClick={this.handleRetry}>
                {t('app.error.retry')}
              </Button>
            ),
            <Button
              type={isChunkError ? 'default' : 'primary'}
              key="reload"
              onClick={this.handleReload}
            >
              {t('app.error.reload')}
            </Button>,
            <Button href="/" key="home">
              {t('app.error.home')}
            </Button>,
          ].filter(Boolean)}
        />
      </Card>
    );
  }
}

const TranslatedErrorBoundary: React.ComponentType<{ children: React.ReactNode }> =
  withTranslation()(ErrorBoundary);

export default TranslatedErrorBoundary;
