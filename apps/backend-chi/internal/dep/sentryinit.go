package dep

import (
	"github.com/getsentry/sentry-go"
)

func InitSentry(appMode AppModeType, dsn string) error {
	if appMode != EnvProd || dsn == "" {
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
		// Enable printing of SDK debug messages.

		// Useful when getting started or trying to figure something out.
		Debug: true,
		// Adds request headers and IP for users,
		// visit: https://docs.sentry.io/platforms/go/data-management/data-collected/ for more info
		AttachStacktrace: true,
		EnableLogs:       true,
	})

	return err
}
