// Package trusted contains check agent subnet middleware
package trusted

import (
	"fmt"
	"net"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

const IPXRealHeader = "X-Real-IP"

func NewTrustedSubnetCheckMW(trustedSubnet string) (middleware.Middleware, error) {

	if trustedSubnet == "" {
		return func(next http.Handler) http.Handler {
			return next
		}, nil
	} else {

		_, n, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, err
		}

		action := domain.GetAction(1)
		logger := domain.GetMainLogger()
		logger.Infow(action, "msg", fmt.Sprintf("trusted subnet %s used", trustedSubnet))

		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				log := domain.GetCtxLogger(req.Context())

				ipHeaderValue := req.Header.Get(IPXRealHeader)
				if ipHeaderValue == "" {
					errMsg := fmt.Sprintf("header %s not specifier", IPXRealHeader)
					log.Errorw("TrustedSubnetCheckMW", "err", errMsg)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				ipAddr := net.ParseIP(ipHeaderValue)
				if ipAddr == nil {
					errMsg := fmt.Sprintf("header %s value %s is not parsable", IPXRealHeader, ipHeaderValue)
					log.Errorw("TrustedSubnetCheckMW", "err", errMsg)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				if !n.Contains(ipAddr) {
					errMsg := fmt.Sprintf("ip %s is not in trusted subnet %s", ipHeaderValue, trustedSubnet)
					log.Errorw("TrustedSubnetCheckMW", "err", errMsg)
					w.WriteHeader(http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, req)
			})
		}, nil
	}
}
