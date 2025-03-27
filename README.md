# grpc_ui_tool
A UI tool using Fyne and Go for connecting to grpc servers, reading protobuf files, and making client calls to the servers.

This is a simple tool that allows you to connect to GRPC servers, read protobuf files, and send client requests.

You can save server connection details to be opened again later for ease of use - please use the .gtserver extension.
Import paths can be saved - but will always be saved as "imports.gtimport" in the same directory as the open protobuf file.

A Few Current Limitations:
* Certs/TLS configurations are unimplemented and it uses insecure tls currently
* Map Types are unimplemented in the input UI
* Cardinality support other than Unary is unimplemented
