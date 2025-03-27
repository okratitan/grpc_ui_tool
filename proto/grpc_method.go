package proto

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetMethods returns a list of grpc methods associated with a grpc service
func (gcd *GrpcConnection) GetMethods(service string) ([]string, error) {
	var methods []string
	gcd.FileRegistry.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		svs := fd.Services()
		for i := 0; i < svs.Len(); i++ {
			if service == string(svs.Get(i).FullName()) {
				m := svs.Get(i).Methods()
				for j := 0; j < m.Len(); j++ {
					methods = append(methods, string(m.Get(j).Name()))
				}
			}
		}
		return true
	})

	sort.Slice(methods, func(i, j int) bool {
		return methods[i] < methods[j]
	})
	return methods, nil
}

func (gcd *GrpcConnection) getMethodDesc(methodName string) (protoreflect.MethodDescriptor, error) {
	desc, err := gcd.FileRegistry.FindDescriptorByName(protoreflect.FullName(methodName))
	if err != nil {
		return nil, err
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("could not find descriptor for method %s", methodName)
	}

	return methodDesc, nil
}
