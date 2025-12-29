generate_proto_files:
	 @for proto_file in proto/*.proto; do \
	 	protoc --proto_path=proto \
		--go_out=paths=source_relative:./protobuf \
		--go-grpc_out=paths=source_relative:./protobuf \
		$$proto_file; \
	 done
