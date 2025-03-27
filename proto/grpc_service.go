package proto

import (
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetServices returns a list of grpc services provided by proto files
func (gcd *GrpcConnection) GetServices() ([]string, error) {
	var services []string
	gcd.FileRegistry.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		svs := fd.Services()
		for i := 0; i < svs.Len(); i++ {
			services = append(services, string(svs.Get(i).FullName()))
		}
		return true
	})

	sort.Slice(services, func(i, j int) bool {
		return services[i] < services[j]
	})
	return services, nil
}
