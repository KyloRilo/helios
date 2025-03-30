// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v3.21.12
// source: pkg/service/api/proto/leader.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BuildReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Image         string                 `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	BuildDir      string                 `protobuf:"bytes,2,opt,name=buildDir,proto3" json:"buildDir,omitempty"`
	Dockerfile    string                 `protobuf:"bytes,3,opt,name=dockerfile,proto3" json:"dockerfile,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BuildReq) Reset() {
	*x = BuildReq{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BuildReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildReq) ProtoMessage() {}

func (x *BuildReq) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildReq.ProtoReflect.Descriptor instead.
func (*BuildReq) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{0}
}

func (x *BuildReq) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

func (x *BuildReq) GetBuildDir() string {
	if x != nil {
		return x.BuildDir
	}
	return ""
}

func (x *BuildReq) GetDockerfile() string {
	if x != nil {
		return x.Dockerfile
	}
	return ""
}

type BuildResp struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Code          int32                  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BuildResp) Reset() {
	*x = BuildResp{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BuildResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildResp) ProtoMessage() {}

func (x *BuildResp) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildResp.ProtoReflect.Descriptor instead.
func (*BuildResp) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{1}
}

func (x *BuildResp) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

type CreateReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Image         string                 `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateReq) Reset() {
	*x = CreateReq{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReq) ProtoMessage() {}

func (x *CreateReq) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReq.ProtoReflect.Descriptor instead.
func (*CreateReq) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{2}
}

func (x *CreateReq) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

type CreateResp struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Code          int32                  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Id            int32                  `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateResp) Reset() {
	*x = CreateResp{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateResp) ProtoMessage() {}

func (x *CreateResp) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateResp.ProtoReflect.Descriptor instead.
func (*CreateResp) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{3}
}

func (x *CreateResp) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *CreateResp) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type StartReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StartReq) Reset() {
	*x = StartReq{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartReq) ProtoMessage() {}

func (x *StartReq) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartReq.ProtoReflect.Descriptor instead.
func (*StartReq) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{4}
}

func (x *StartReq) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type StartResp struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Code          int32                  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StartResp) Reset() {
	*x = StartResp{}
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartResp) ProtoMessage() {}

func (x *StartResp) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_api_proto_leader_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartResp.ProtoReflect.Descriptor instead.
func (*StartResp) Descriptor() ([]byte, []int) {
	return file_pkg_service_api_proto_leader_proto_rawDescGZIP(), []int{5}
}

func (x *StartResp) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

var File_pkg_service_api_proto_leader_proto protoreflect.FileDescriptor

var file_pkg_service_api_proto_leader_proto_rawDesc = string([]byte{
	0x0a, 0x22, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5c, 0x0a, 0x08, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x44, 0x69, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x44, 0x69, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x6f, 0x63,
	0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64,
	0x6f, 0x63, 0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x22, 0x1f, 0x0a, 0x09, 0x42, 0x75, 0x69,
	0x6c, 0x64, 0x52, 0x65, 0x73, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x22, 0x21, 0x0a, 0x09, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x30, 0x0a,
	0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x63,
	0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x22,
	0x1a, 0x0a, 0x08, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x22, 0x1f, 0x0a, 0x09, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x32, 0xac, 0x01, 0x0a,
	0x06, 0x44, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x31, 0x0a, 0x0a, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x75,
	0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x38, 0x0a, 0x0f, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x12, 0x10, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x1a,
	0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x22, 0x00, 0x12, 0x35, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x43, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x12, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x22, 0x5a, 0x20, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4b, 0x79, 0x6c, 0x6f, 0x52, 0x69,
	0x6c, 0x6f, 0x2f, 0x68, 0x65, 0x6c, 0x69, 0x6f, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_pkg_service_api_proto_leader_proto_rawDescOnce sync.Once
	file_pkg_service_api_proto_leader_proto_rawDescData []byte
)

func file_pkg_service_api_proto_leader_proto_rawDescGZIP() []byte {
	file_pkg_service_api_proto_leader_proto_rawDescOnce.Do(func() {
		file_pkg_service_api_proto_leader_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_pkg_service_api_proto_leader_proto_rawDesc), len(file_pkg_service_api_proto_leader_proto_rawDesc)))
	})
	return file_pkg_service_api_proto_leader_proto_rawDescData
}

var file_pkg_service_api_proto_leader_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_pkg_service_api_proto_leader_proto_goTypes = []any{
	(*BuildReq)(nil),   // 0: proto.BuildReq
	(*BuildResp)(nil),  // 1: proto.BuildResp
	(*CreateReq)(nil),  // 2: proto.CreateReq
	(*CreateResp)(nil), // 3: proto.CreateResp
	(*StartReq)(nil),   // 4: proto.StartReq
	(*StartResp)(nil),  // 5: proto.StartResp
}
var file_pkg_service_api_proto_leader_proto_depIdxs = []int32{
	0, // 0: proto.Docker.BuildImage:input_type -> proto.BuildReq
	2, // 1: proto.Docker.CreateContainer:input_type -> proto.CreateReq
	4, // 2: proto.Docker.StartContainer:input_type -> proto.StartReq
	1, // 3: proto.Docker.BuildImage:output_type -> proto.BuildResp
	3, // 4: proto.Docker.CreateContainer:output_type -> proto.CreateResp
	5, // 5: proto.Docker.StartContainer:output_type -> proto.StartResp
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_service_api_proto_leader_proto_init() }
func file_pkg_service_api_proto_leader_proto_init() {
	if File_pkg_service_api_proto_leader_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_pkg_service_api_proto_leader_proto_rawDesc), len(file_pkg_service_api_proto_leader_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_service_api_proto_leader_proto_goTypes,
		DependencyIndexes: file_pkg_service_api_proto_leader_proto_depIdxs,
		MessageInfos:      file_pkg_service_api_proto_leader_proto_msgTypes,
	}.Build()
	File_pkg_service_api_proto_leader_proto = out.File
	file_pkg_service_api_proto_leader_proto_goTypes = nil
	file_pkg_service_api_proto_leader_proto_depIdxs = nil
}
