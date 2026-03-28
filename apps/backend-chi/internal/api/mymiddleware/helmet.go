//Re-written from https://github.com/gofiber/helmet/blob/v2.2.26/main.go

package mymiddleware

import (
	"fmt"
	"net/http"
	"strings"
)

// Config defines security header options.
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*http.Request) bool
	// XSSProtection
	// Optional. Default value "0".
	XSSProtection string
	// ContentTypeNosniff
	// Optional. Default value "nosniff".
	ContentTypeNosniff string
	// XFrameOptions
	// Optional. Default value "SAMEORIGIN".
	// Possible values: "SAMEORIGIN", "DENY", "ALLOW-FROM uri"
	XFrameOptions string
	// HSTSMaxAge
	// Optional. Default value 0.
	HSTSMaxAge int
	// HSTSExcludeSubdomains
	// Optional. Default value false.
	HSTSExcludeSubdomains bool
	// ContentSecurityPolicy
	// Optional. Default value "".
	ContentSecurityPolicy string
	// CSPReportOnly
	// Optional. Default value false.
	CSPReportOnly bool
	// HSTSPreloadEnabled
	// Optional. Default value false.
	HSTSPreloadEnabled bool
	// ReferrerPolicy
	// Optional. Default value "no-referrer".
	ReferrerPolicy string
	// Permissions-Policy
	// Optional. Default value "".
	PermissionPolicy string
	// Cross-Origin-Embedder-Policy
	// Optional. Default value "require-corp".
	CrossOriginEmbedderPolicy string
	// Cross-Origin-Opener-Policy
	// Optional. Default value "same-origin".
	CrossOriginOpenerPolicy string
	// Cross-Origin-Resource-Policy
	// Optional. Default value "same-origin".
	CrossOriginResourcePolicy string
	// Origin-Agent-Cluster
	// Optional. Default value "?1".
	OriginAgentCluster string
	// X-DNS-Prefetch-Control
	// Optional. Default value "off".
	XDNSPrefetchControl string
	// X-Download-Options
	// Optional. Default value "noopen".
	XDownloadOptions string
	// X-Permitted-Cross-Domain-Policies
	// Optional. Default value "none".
	XPermittedCrossDomain string
}

// Helmet sets common security headers.
func Helmet(config ...Config) func(http.Handler) http.Handler {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.XSSProtection == "" {
		cfg.XSSProtection = "0"
	}
	if cfg.ContentTypeNosniff == "" {
		cfg.ContentTypeNosniff = "nosniff"
	}
	if cfg.XFrameOptions == "" {
		cfg.XFrameOptions = "SAMEORIGIN"
	}
	if cfg.ReferrerPolicy == "" {
		cfg.ReferrerPolicy = "no-referrer"
	}
	if cfg.CrossOriginEmbedderPolicy == "" {
		cfg.CrossOriginEmbedderPolicy = "require-corp"
	}
	if cfg.CrossOriginOpenerPolicy == "" {
		cfg.CrossOriginOpenerPolicy = "same-origin"
	}
	if cfg.CrossOriginResourcePolicy == "" {
		cfg.CrossOriginResourcePolicy = "same-origin"
	}
	if cfg.OriginAgentCluster == "" {
		cfg.OriginAgentCluster = "?1"
	}
	if cfg.XDNSPrefetchControl == "" {
		cfg.XDNSPrefetchControl = "off"
	}
	if cfg.XDownloadOptions == "" {
		cfg.XDownloadOptions = "noopen"
	}
	if cfg.XPermittedCrossDomain == "" {
		cfg.XPermittedCrossDomain = "none"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Filter != nil && cfg.Filter(r) {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}
			if cfg.ContentTypeNosniff != "" {
				w.Header().Set("X-Content-Type-Options", cfg.ContentTypeNosniff)
			}
			if cfg.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}
			if cfg.CrossOriginEmbedderPolicy != "" {
				w.Header().Set("Cross-Origin-Embedder-Policy", cfg.CrossOriginEmbedderPolicy)
			}
			if cfg.CrossOriginOpenerPolicy != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
			}
			if cfg.CrossOriginResourcePolicy != "" {
				w.Header().Set("Cross-Origin-Resource-Policy", cfg.CrossOriginResourcePolicy)
			}
			if cfg.OriginAgentCluster != "" {
				w.Header().Set("Origin-Agent-Cluster", cfg.OriginAgentCluster)
			}
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}
			if cfg.XDNSPrefetchControl != "" {
				w.Header().Set("X-DNS-Prefetch-Control", cfg.XDNSPrefetchControl)
			}
			if cfg.XDownloadOptions != "" {
				w.Header().Set("X-Download-Options", cfg.XDownloadOptions)
			}
			if cfg.XPermittedCrossDomain != "" {
				w.Header().Set("X-Permitted-Cross-Domain-Policies", cfg.XPermittedCrossDomain)
			}

			if isHTTPS(r) && cfg.HSTSMaxAge != 0 {
				subdomains := ""
				if !cfg.HSTSExcludeSubdomains {
					subdomains = "; includeSubDomains"
				}
				if cfg.HSTSPreloadEnabled {
					subdomains = fmt.Sprintf("%s; preload", subdomains)
				}
				w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d%s", cfg.HSTSMaxAge, subdomains))
			}

			if cfg.ContentSecurityPolicy != "" {
				if cfg.CSPReportOnly {
					w.Header().Set("Content-Security-Policy-Report-Only", cfg.ContentSecurityPolicy)
				} else {
					w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
				}
			}

			if cfg.PermissionPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
