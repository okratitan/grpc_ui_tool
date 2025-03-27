package proto

import (
	"context"
	"crypto/tls"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GrpcConnection struct {
	Hostname     string
	Port         string
	Metadata     map[string]string
	FileRegistry *protoregistry.Files
}

func NewGrpcConnection() *GrpcConnection {
	return &GrpcConnection{}
}

// SetConnectionDetails sets the grpc server details to be used for a client connection to the server
func (gcd *GrpcConnection) SetConnectionDetails(hostname string, port string, metadata map[string]string) {
	gcd.Hostname = hostname
	gcd.Port = port
	gcd.Metadata = metadata
}

// LoadRegistry takes a .proto file and loads the file and associated imports/paths into a grpc file registry
func (gcd *GrpcConnection) LoadRegistry(importPaths []string, protoFile string) error {

	f, err := protoparse.ResolveFilenames(importPaths, protoFile)
	if err != nil {
		return err
	}

	parser := protoparse.Parser{
		ImportPaths:      importPaths,
		InferImportPaths: true,
	}

	fds, err := parser.ParseFiles(f...)
	if err != nil {
		return err
	}

	fdSet := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]struct{})

	for _, fd := range fds {
		fdSet.File = append(fdSet.File, gcd.walkFileDescriptors(seen, fd)...)
	}

	gcd.FileRegistry, err = protodesc.NewFiles(fdSet)
	if err != nil {
		return err
	}

	return nil
}

// Send will connect to the grpc server and send a grpc request
func (gcd *GrpcConnection) Send(serviceName string, methodName string, jsonRequest string) (string, error) {
	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
		grpc.WithUserAgent("grpc-tool/1.0"),
	}

	conn, err := grpc.NewClient(gcd.Hostname+":"+gcd.Port, opts...)
	if err != nil {
		return "", err
	}

	messageDesc, err := gcd.getMethodDesc(serviceName + "." + methodName)
	if err != nil {
		return "", err
	}

	req := dynamicpb.NewMessage(messageDesc.Input())
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(jsonRequest), req); err != nil {
		return "", err
	}
	ctx := metadata.NewOutgoingContext(context.Background(), nil)
	for key, value := range gcd.Metadata {
		ctx = metadata.AppendToOutgoingContext(ctx, key, value)
	}

	resp := dynamicpb.NewMessage(messageDesc.Output())
	conn.Connect()
	formatMethodName := "/" + serviceName + "/" + string(messageDesc.Name())
	err = conn.Invoke(ctx, formatMethodName, req, resp)
	if err != nil {
		return "", err
	}

	prettified, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(resp)
	if err != nil {
		return "", err
	}

	err = conn.Close()
	if err != nil {
		return string(prettified), err
	}

	return string(prettified), nil
}

func (gcd *GrpcConnection) walkFileDescriptors(seen map[string]struct{}, fd *desc.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	var fds []*descriptorpb.FileDescriptorProto

	if _, ok := seen[fd.GetName()]; ok {
		return fds
	}
	seen[fd.GetName()] = struct{}{}
	fds = append(fds, fd.AsFileDescriptorProto())

	for _, dep := range fd.GetDependencies() {
		deps := gcd.walkFileDescriptors(seen, dep)
		fds = append(fds, deps...)
	}

	return fds
}
