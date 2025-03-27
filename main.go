package main

import (
	"grpc_ui_tool/proto"
	"grpc_ui_tool/ui"
)

var grpcConn *proto.GrpcConnection

func main() {
	grpcConn = proto.NewGrpcConnection()
	_ = ui.CreateUI(grpcConn)

}
