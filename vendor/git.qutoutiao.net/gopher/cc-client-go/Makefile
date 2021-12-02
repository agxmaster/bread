IDL := cc-idl
PROTOC := protoc
PROTOC_INCLUDES := \
	-I /usr/local/include
PROTO_MAPPINGS := ""
.PHONY: proto
proto:
	$(PROTOC) \
		$(PROTOC_INCLUDES) \
		-I $(IDL)/admin \
		--go_out=plugins=grpc,$(PROTO_MAPPINGS):$(PWD)/proto-gen/admin_sdk  $(IDL)/admin/admin-sdk.proto