package company

import (
	"fmt"

	"google.golang.org/grpc"

	tracing "v2.staffjoy.com/tracing"
	otgrpc "github.com/opentracing-contrib/go-grpc"
)

var company_client CompanyServiceClient
var empty_fun = func() error {
	return nil
}

// NewClient returns a gRPC client for interacting with the company.
// After calling it, run a defer close on the close function
func NewClient(serviceName string) (CompanyServiceClient, func() error, error) {
	if company_client == nil {
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
		company_client = NewCompanyServiceClient(conn)
	}
	return company_client, empty_fun, nil
}
