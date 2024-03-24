package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	// cHttp is a map that contains custom mappings for codes to http codes
	cHttp = make(map[Code]int)
	// cGrpc is a map that contains custom mappings for codes to grpc codes
	cGrpc = make(map[Code]codes.Code)
	// cDesc is a map that contains custom mappings for codes to their descriptions
	cDesc = make(map[Code]string)
)

// Code is the type that represents an error code.
// It can map to HTTP and gRPC codes.
// In order to properly work with custom codes or code overrides:
// use the RegisterCode function after creating your new Code instance.
type Code int

const (
	// Unknown is the default error code.
	Unknown Code = iota

	// Canceled indicates the operation was canceled or unavailable because it was cancelled.
	Canceled

	// InvalidArgument is used when the client sends invalid arguments.
	InvalidArgument

	// DeadlineExceeded means operation expired before completion. This doesn't necessarily mean that the operation failed.
	// It is possible that the operation succeeded but the deadline was exceeded.
	DeadlineExceeded

	// Unauthenticated is used when the client is not authenticated.
	Unauthenticated

	// NotFound is used when the requested resource is not found.
	NotFound

	// AlreadyExists is used when the resource already exists.
	AlreadyExists

	// Forbidden is used when the client is not authorized to perform the requested operation.
	Forbidden

	// ResourceExhausted is used when the client has exhausted some resource.
	ResourceExhausted

	// FailedPrecondition is used when the client sends a request that is not allowed in the current state.
	FailedPrecondition

	// Aborted is used when the client sends a request that cannot be completed due to a conflict e.g. a concurrency issue.
	Aborted

	// OutOfRange means that the operation was attempted past the valid range.
	OutOfRange

	// Internal is used when an internal error occurs.
	Internal

	// Unavailable is used when the service cannot be reached due to some network issues.
	Unavailable

	// DataLoss is used when the service has lost some data.
	DataLoss
)

// CodeSize is the number of codes defined in the STL library.
// All codes defined by default are mapped from 0 to CodeSize - 1.
// CodeSize can be useful when creating additional codes for example:
//
//	const (
//		// MyCode is a custom code.
//		MyCode Code = errs.CodeSize + iota // = 15
//	 	ExtraCode // = 16
//	)
const CodeSize = 15

// String returns the string representation of the code.
func (c Code) String() string {
	if x, ok := cDesc[c]; ok {
		return x
	}
	return codeNames[c]
}

// MarshalJSON implements the json.Marshaler interface and defines how a Code
// should be marshaled to JSON. By default, it marshals to a string representation defined by String function.
func (c Code) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.String() + `"`), nil
}

// HTTP returns the HTTP code that is mapped to the code.
func (c Code) HTTP() int {
	if x, ok := cHttp[c]; ok {
		return x
	}
	return httpCodes[c]
}

// GRPC returns the gPRC code that is mapped to the code.
func (c Code) GRPC() codes.Code {
	if x, ok := cGrpc[c]; ok {
		return x
	}
	return grpcCodes[c]
}

// RegisterCode registers a new code OR overrides an existing one.
func RegisterCode(c Code, HTTP int, GRPC codes.Code, desc string) {
	cHttp[c] = HTTP
	cGrpc[c] = GRPC
	cDesc[c] = desc
}

// UnregisterCode unregisters the custom implementation or override of a code
// provided from the RegisterCode function.
// When a code is unregistered, UnregisterCode is a no-op.
func UnregisterCode(c Code) {
	delete(cHttp, c)
	delete(cGrpc, c)
	delete(cDesc, c)
}

// IsRegistered returns true if a custom implementation or override is being used for the code.
func IsRegistered(c Code) bool {
	_, ok := cHttp[c]
	return ok
}

// ClearCodeRegister removes all registration made
// with the function RegisterCode
func ClearCodeRegister() {
	cHttp = make(map[Code]int)
	cGrpc = make(map[Code]codes.Code)
	cDesc = make(map[Code]string)
}

// httpCodes is an array that contains DEFAULT mappings for
// codes to http codes
var httpCodes = [...]int{
	Unknown:            http.StatusInternalServerError,
	DataLoss:           http.StatusInternalServerError,
	DeadlineExceeded:   http.StatusGatewayTimeout,
	FailedPrecondition: http.StatusPreconditionFailed,
	ResourceExhausted:  http.StatusTooManyRequests,
	Canceled:           499,
	InvalidArgument:    http.StatusBadRequest,
	Unauthenticated:    http.StatusUnauthorized,
	Forbidden:          http.StatusForbidden,
	NotFound:           http.StatusNotFound,
	AlreadyExists:      http.StatusConflict,
	Internal:           http.StatusInternalServerError,
	OutOfRange:         http.StatusRequestedRangeNotSatisfiable,
	Aborted:            http.StatusConflict,
	Unavailable:        http.StatusServiceUnavailable,
}

// codeNames is an array that contains DEFAULT string descriptions of codes
var codeNames = [...]string{
	Unknown:            "unknown",
	DataLoss:           "data_loss",
	DeadlineExceeded:   "deadline_exceeded",
	FailedPrecondition: "failed_precondition",
	ResourceExhausted:  "resource_exhausted",
	Canceled:           "canceled",
	InvalidArgument:    "invalid_argument",
	Unauthenticated:    "unauthenticated",
	Forbidden:          "forbidden",
	NotFound:           "not_found",
	AlreadyExists:      "already_exists",
	Internal:           "internal",
	OutOfRange:         "out_of_range",
	Aborted:            "aborted",
	Unavailable:        "unavailable",
}

// grpcCodes is an array that contains default 1 to 1 mapping of codes
// to grpc codes.
var grpcCodes = [...]codes.Code{
	DataLoss:           codes.DataLoss,
	Unknown:            codes.Unknown,
	InvalidArgument:    codes.InvalidArgument,
	DeadlineExceeded:   codes.DeadlineExceeded,
	FailedPrecondition: codes.FailedPrecondition,
	ResourceExhausted:  codes.ResourceExhausted,
	Canceled:           codes.Canceled,
	Unauthenticated:    codes.Unauthenticated,
	Forbidden:          codes.PermissionDenied,
	NotFound:           codes.NotFound,
	AlreadyExists:      codes.AlreadyExists,
	Internal:           codes.Internal,
	Aborted:            codes.Aborted,
	OutOfRange:         codes.OutOfRange,
	Unavailable:        codes.Unavailable,
}
