import * as Sentry from '@sentry/react';

interface Logger {
  info: (...args: unknown[]) => void;
  warn: (...args: unknown[]) => void;
  debug: (...args: unknown[]) => void;
  error: (...args: unknown[]) => void;
}

const consoleLogger: Logger = {
  info: (...args) => {
    console.info('[INFO]', ...args);
  },
  warn: (...args) => {
    console.warn('[WARN]', ...args);
  },
  debug: (...args) => {
    console.debug('[DEBUG]', ...args);
  },
  error: (...args) => {
    const maybeError = args.find((arg) => arg instanceof Error);
    if (maybeError) Sentry.captureException(maybeError);

    console.error('[ERROR]', ...args);
  },
};

const logger = consoleLogger;

export default logger;
