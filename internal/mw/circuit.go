package mw

import (
	"context"

	"github.com/catalystgo/catalystgo/errors"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
)

func UnaryCircuitBreakerInterceptor(settings gobreaker.Settings) grpc.UnaryServerInterceptor {
	// settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
	// 	return counts.ConsecutiveFailures > 5
	// }
	// settings.IsSuccessful = func(err error) bool {
	// 	if err == nil {
	// 		return true
	// 	}

	// 	code := status.Code(err)

	// 	// TODO: Make the error list configurable
	// 	// Do not trip the circuit breaker for internal errors
	// 	// or resource exhausted errors (e.g. too many requests) or unknown errors
	// 	// Since these errors are not due to the service being unhealthy or unavailable but
	// 	// due to the client sending bad requests or the server being overloaded
	// 	if code == codes.Internal || code == codes.Unknown || code == codes.ResourceExhausted {
	// 		return false
	// 	}

	// 	return true
	// }

	cb := gobreaker.NewCircuitBreaker[interface{}](settings)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Wrap the handler with the circuit breaker's Execute method
		// This ensures that the handler is only called if the circuit is closed
		output, err := cb.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})

		if err != nil {
			// Check if the error is a circuit breaker open error
			if errors.Is(err, gobreaker.ErrOpenState) {
				err = errors.
					Newf("circuit breaker is open.").
					Code(errors.ResourceExhausted).
					Op(info.FullMethod)
			}
			return nil, err
		}
		return output, nil
	}
}

func StreamCircuitBreakerInterceptor(settings gobreaker.Settings) grpc.StreamServerInterceptor {
	cb := gobreaker.NewCircuitBreaker[interface{}](settings)

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Wrap the handler with the circuit breaker's Execute method
		// This ensures that the handler is only called if the circuit is closed
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, handler(srv, stream)
		})

		if err != nil {
			// Check if the error is a circuit breaker open error
			if errors.Is(err, gobreaker.ErrOpenState) {
				err = errors.
					Newf("circuit breaker is open.").
					Code(errors.ResourceExhausted).
					Op(info.FullMethod)
			}
			return err
		}
		return nil
	}
}
