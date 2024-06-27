package catalystgo

import (
	"context"

	"github.com/catalystgo/logger/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type (
	Service interface {
		GetDescription() ServiceDesc
	}

	ServiceDesc interface {
		RegisterGRPC(s *grpc.Server)
		RegisterHTTP(ctx context.Context, mux *runtime.ServeMux) error
		SwaggerJSON() []byte
		WithHTTPUnaryInterceptor(i grpc.UnaryServerInterceptor)
	}
)

type compoundServiceDesc struct {
	desc []ServiceDesc
}

func NewCompoundServiceDesc(desc ...ServiceDesc) ServiceDesc {
	return &compoundServiceDesc{desc: desc}
}

func (c *compoundServiceDesc) RegisterGRPC(s *grpc.Server) {
	for _, d := range c.desc {
		d.RegisterGRPC(s)
	}
}

func (c *compoundServiceDesc) RegisterHTTP(ctx context.Context, mux *runtime.ServeMux) error {
	for _, d := range c.desc {
		if err := d.RegisterHTTP(ctx, mux); err != nil {
			logger.Errorf(ctx, "register http: %v", err)
		}
	}
	return nil
}

func (c *compoundServiceDesc) SwaggerJSON() []byte {
	return c.desc[len(c.desc)-1].SwaggerJSON() // TODO: merge swagger defs from multiple services (currently only the last one is used)
}

func (c *compoundServiceDesc) WithHTTPUnaryInterceptor(i grpc.UnaryServerInterceptor) {
	for _, d := range c.desc {
		d.WithHTTPUnaryInterceptor(i)
	}
}
