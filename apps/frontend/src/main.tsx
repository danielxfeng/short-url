import './instrument';
import { createRoot } from 'react-dom/client';
import './index.css';
import * as Sentry from '@sentry/react';
import App from './App.tsx';

const container = document.getElementById('root');

if (!container) {
  throw new Error('Failed to find the root element');
}

const root = createRoot(container, {
  // Callback called when an error is thrown and not caught by an ErrorBoundary.
  onUncaughtError: Sentry.reactErrorHandler((error, errorInfo) => {
    console.warn('Uncaught error', error, errorInfo.componentStack);
  }),
  // Callback called when React catches an error in an ErrorBoundary.
  onCaughtError: Sentry.reactErrorHandler(),
  // Callback called when React automatically recovers from errors.
  onRecoverableError: Sentry.reactErrorHandler(),
});

root.render(
  <Sentry.ErrorBoundary fallback={<div>Something went wrong.</div>}>
    <App />
  </Sentry.ErrorBoundary>,
);
