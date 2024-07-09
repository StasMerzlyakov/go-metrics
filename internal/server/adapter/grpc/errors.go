package grpc

import (
	"errors"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"

	gp "google.golang.org/grpc/codes"
)

func MapDomainErrorToGRPCCodeErr(err error) gp.Code {

	if errors.Is(err, domain.ErrMediaType) {
		return gp.Unknown
	}

	if errors.Is(err, domain.ErrDataFormat) {
		return gp.InvalidArgument
	}

	if errors.Is(err, domain.ErrDataDigestMismath) {
		return gp.Unauthenticated
	}

	if errors.Is(err, domain.ErrServerInternal) {
		return gp.Aborted
	}

	if errors.Is(err, domain.ErrDBConnection) {
		return gp.FailedPrecondition
	}

	if errors.Is(err, domain.ErrNotFound) {
		return gp.Unavailable
	}

	return gp.Aborted
}
