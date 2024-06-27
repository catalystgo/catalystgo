package mw

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimiter defines the rate limiter struct.
type RateLimiter struct {
	mu           sync.RWMutex
	defaultLimit *rate.Limiter

	limits map[string]*rate.Limiter
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(defaultRate rate.Limit, defaultBurst int) *RateLimiter {
	return &RateLimiter{
		defaultLimit: rate.NewLimiter(defaultRate, defaultBurst),
		limits:       make(map[string]*rate.Limiter),
	}
}

// SetLimit sets the rate limit for a specific method.
func (rl *RateLimiter) SetLimit(method string, r rate.Limit, b int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limits[method] = rate.NewLimiter(rate.Limit(r), b)
}

// GetLimit gets the rate limit for a specific method, or returns the default limit if not set.
func (rl *RateLimiter) GetLimit(method string) *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if limiter, ok := rl.limits[method]; ok {
		return limiter
	}
	return rl.defaultLimit
}

// ChangeLimit changes the rate limit for a specific method.
func (rl *RateLimiter) ChangeLimit(method string, r rate.Limit, b int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if limiter, ok := rl.limits[method]; ok {
		limiter.SetLimit(rate.Limit(r))
		limiter.SetBurst(b)
	} else {
		rl.limits[method] = rate.NewLimiter(rate.Limit(r), b)
	}
}

// UnaryInterceptor returns a gRPC unary server interceptor for rate limiting.
func (rl *RateLimiter) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		limiter := rl.GetLimit(info.FullMethod)
		if limiter.Allow() {
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}
}

// StreamInterceptor returns a gRPC stream server interceptor for rate limiting.
func (rl *RateLimiter) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		limiter := rl.GetLimit(info.FullMethod)
		if limiter.Allow() {
			return handler(srv, stream)
		}
		return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}
}
