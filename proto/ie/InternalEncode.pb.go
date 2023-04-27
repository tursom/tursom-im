// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.22.3
// source: InternalEncode.proto

package ie

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type KVStoreKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Content:
	//
	//	*KVStoreKey_System
	Content isKVStoreKey_Content `protobuf_oneof:"content"`
}

func (x *KVStoreKey) Reset() {
	*x = KVStoreKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_InternalEncode_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVStoreKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVStoreKey) ProtoMessage() {}

func (x *KVStoreKey) ProtoReflect() protoreflect.Message {
	mi := &file_InternalEncode_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVStoreKey.ProtoReflect.Descriptor instead.
func (*KVStoreKey) Descriptor() ([]byte, []int) {
	return file_InternalEncode_proto_rawDescGZIP(), []int{0}
}

func (m *KVStoreKey) GetContent() isKVStoreKey_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (x *KVStoreKey) GetSystem() string {
	if x, ok := x.GetContent().(*KVStoreKey_System); ok {
		return x.System
	}
	return ""
}

type isKVStoreKey_Content interface {
	isKVStoreKey_Content()
}

type KVStoreKey_System struct {
	System string `protobuf:"bytes,1,opt,name=system,proto3,oneof"`
}

func (*KVStoreKey_System) isKVStoreKey_Content() {}

var File_InternalEncode_proto protoreflect.FileDescriptor

var file_InternalEncode_proto_rawDesc = []byte{
	0x0a, 0x14, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x31, 0x0a, 0x0a, 0x4b, 0x56, 0x53, 0x74, 0x6f, 0x72,
	0x65, 0x4b, 0x65, 0x79, 0x12, 0x18, 0x0a, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x42, 0x09,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x42, 0x08, 0x48, 0x01, 0x5a, 0x04, 0x2e,
	0x2f, 0x69, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_InternalEncode_proto_rawDescOnce sync.Once
	file_InternalEncode_proto_rawDescData = file_InternalEncode_proto_rawDesc
)

func file_InternalEncode_proto_rawDescGZIP() []byte {
	file_InternalEncode_proto_rawDescOnce.Do(func() {
		file_InternalEncode_proto_rawDescData = protoimpl.X.CompressGZIP(file_InternalEncode_proto_rawDescData)
	})
	return file_InternalEncode_proto_rawDescData
}

var file_InternalEncode_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_InternalEncode_proto_goTypes = []interface{}{
	(*KVStoreKey)(nil), // 0: KVStoreKey
}
var file_InternalEncode_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_InternalEncode_proto_init() }
func file_InternalEncode_proto_init() {
	if File_InternalEncode_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_InternalEncode_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVStoreKey); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_InternalEncode_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*KVStoreKey_System)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_InternalEncode_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_InternalEncode_proto_goTypes,
		DependencyIndexes: file_InternalEncode_proto_depIdxs,
		MessageInfos:      file_InternalEncode_proto_msgTypes,
	}.Build()
	File_InternalEncode_proto = out.File
	file_InternalEncode_proto_rawDesc = nil
	file_InternalEncode_proto_goTypes = nil
	file_InternalEncode_proto_depIdxs = nil
}