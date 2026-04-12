import './instrument';
import { createRoot } from 'react-dom/client';
import './index.css';
import * as Sentry from '@sentry/react';
import App from './App.tsx';
import { BrowserRouter } from 'react-router';
import { Toaster } from './components/ui/sonner.tsx';
import { TooltipProvider } from './components/ui/tooltip.tsx';
import { QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { toast } from 'sonner';
import logger from './lib/logger.ts';
import ErrorPage from './components/ErrorPage.tsx';

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

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error, query) => {
      if (query.meta?.errorMessage) {
        const errorMessage =
          typeof query.meta.errorMessage === 'string'
            ? query.meta.errorMessage
            : 'An error occurred, please try again';

        logger.error('Query error', { error, queryKey: query.queryKey, errorMessage });
        toast.error(errorMessage);
      }
    },
  }),
});

root.render(
  <Sentry.ErrorBoundary fallback={<ErrorPage />}>
    <BrowserRouter>
      <TooltipProvider>
        <QueryClientProvider client={queryClient}>
          <App />
        </QueryClientProvider>
      </TooltipProvider>
      <Toaster richColors={true} position='top-center' />
    </BrowserRouter>
  </Sentry.ErrorBoundary>,
);
