package account

import (
	"fmt"

	"google.golang.org/grpc"

	tracing "v2.staffjoy.com/tracing"
	otgrpc "github.com/opentracing-contrib/go-grpc"
)

var account_client AccountServiceClient
var empty_fun = func() error {
	return nil
}

// NewClient returns a gRPC client for interacting with the account.
// After calling it, run a defer close on the close function
func NewClient(serviceName string) (AccountServiceClient, func() error, error) {
	if account_client == nil {
		tracer := tracing.GetTracer()
		conn, err := grpc.Dial(Endpoint,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(tracer)),
			grpc.WithStreamInterceptor(
				otgrpc.OpenTracingStreamClientInterceptor(tracer)))
		if err != nil {
			return nil, nil, fmt.Errorf("did not connect: %v", err)
		}
		account_client = NewAccountServiceClient(conn)
	}
	return account_client, empty_fun, nil
}
