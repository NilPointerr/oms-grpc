# generate_proto_files:
# 	 protoc --proto_path=proto \
#      --proto_path=. \
# 	 --go_out=paths=source_relative:./protobuf \
# 	 --go-grpc_out=paths=source_relative:./protobuf \
# 	 proto/oms_service.proto


generate_proto_files:
	 @for proto_file in proto/*.proto; do \
	 	protoc --proto_path=proto \
		--go_out=paths=source_relative:./protobuf \
		--go-grpc_out=paths=source_relative:./protobuf \
		$$proto_file; \
	 done


# generate_proto_files:
# 	 @protoc --proto_path=proto \
# 		--go_out=paths=source_relative:./protobuf \
# 		--go-grpc_out=paths=source_relative:./protobuf \
# 		/oms_employee.proto \
# 		/oms_service.proto
