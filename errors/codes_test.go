package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestCodeHTTP(t *testing.T) {
	for i := 0; i < CodeSize; i++ {
		testName := fmt.Sprintf("%s should not be zero", Code(i).String())
		t.Run(testName, func(t *testing.T) {
			if Code(i).HTTP() == 0 {
				t.Errorf("Code(%d).HTTP() == 0", i)
			}
		})
	}
}

func TestCodeGRPC(t *testing.T) {
	for i := 0; i < CodeSize; i++ {
		testName := fmt.Sprintf("%s should not be zero", Code(i).String())
		t.Run(testName, func(t *testing.T) {
			if Code(i).GRPC() == codes.OK {
				t.Errorf("Code(%d).GRPC() == 0", i)
			}
		})
	}
}

func TestCodeString(t *testing.T) {
	for i := 0; i < CodeSize; i++ {
		testName := fmt.Sprintf("%s should not be empty", Code(i).String())
		t.Run(testName, func(t *testing.T) {
			if Code(i).String() == "" {
				t.Errorf("Code(%d).String() == \"\"", i)
			}
		})
	}
}

func TestRegisterCode(t *testing.T) {
	type args struct {
		c    Code
		http int
		grpc codes.Code
		desc string
	}
	testCases := []struct {
		name      string
		args      args
		shouldErr bool
	}{
		{
			name: "registering entirely new codes",
			args: args{
				c:    Code(100),
				http: 100,
				grpc: codes.Code(100),
				desc: "test",
			},
		},
		{
			name: "overriding existing codes",
			args: args{
				c:    NotFound,
				http: 401,
				grpc: codes.Unknown,
				desc: "this is not found",
			},
		},
	}
	for _, tt := range testCases {
		t.Run("Register "+tt.name, func(t *testing.T) {
			// test register
			RegisterCode(tt.args.c, tt.args.http, tt.args.grpc, tt.args.desc)
			assert.Equal(t, tt.args.http, tt.args.c.HTTP())
			assert.Equal(t, tt.args.grpc, tt.args.c.GRPC())
			assert.Equal(t, tt.args.desc, tt.args.c.String())

			// test register check
			assert.True(t, IsRegistered(tt.args.c))
		})
		t.Run("Unregister "+tt.name, func(t *testing.T) {
			// unregister and check
			UnregisterCode(tt.args.c)
			assert.False(t, IsRegistered(tt.args.c))
		})
	}

	// test ClearCodeRegister
	t.Run("ClearCodeRegister", func(t *testing.T) {
		// given
		for _, tt := range testCases {
			RegisterCode(tt.args.c, tt.args.http, tt.args.grpc, tt.args.desc)
		}

		// when
		ClearCodeRegister()

		// then
		for _, tt := range testCases {
			assert.False(t, IsRegistered(tt.args.c))
		}
	})
}
