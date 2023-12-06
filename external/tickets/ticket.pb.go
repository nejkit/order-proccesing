// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.6
// source: ticket.proto

package tickets

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

type TicketState int32

const (
	TicketState_TICKET_STATE_NEW             TicketState = 0
	TicketState_TICKET_STATE_WAIT_PROCESSING TicketState = 1
	TicketState_TICKET_STATE_PROCESSING      TicketState = 2
	TicketState_TICKET_STATE_DONE            TicketState = 3
)

// Enum value maps for TicketState.
var (
	TicketState_name = map[int32]string{
		0: "TICKET_STATE_NEW",
		1: "TICKET_STATE_WAIT_PROCESSING",
		2: "TICKET_STATE_PROCESSING",
		3: "TICKET_STATE_DONE",
	}
	TicketState_value = map[string]int32{
		"TICKET_STATE_NEW":             0,
		"TICKET_STATE_WAIT_PROCESSING": 1,
		"TICKET_STATE_PROCESSING":      2,
		"TICKET_STATE_DONE":            3,
	}
)

func (x TicketState) Enum() *TicketState {
	p := new(TicketState)
	*p = x
	return p
}

func (x TicketState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TicketState) Descriptor() protoreflect.EnumDescriptor {
	return file_ticket_proto_enumTypes[0].Descriptor()
}

func (TicketState) Type() protoreflect.EnumType {
	return &file_ticket_proto_enumTypes[0]
}

func (x TicketState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TicketState.Descriptor instead.
func (TicketState) EnumDescriptor() ([]byte, []int) {
	return file_ticket_proto_rawDescGZIP(), []int{0}
}

type OperationType int32

const (
	OperationType_OPERATION_TYPE_CREATE_ORDER          OperationType = 0
	OperationType_OPERATION_TYPE_LOCK_BALANCE          OperationType = 1
	OperationType_OPERATION_TYPE_APPROVE_CREATION      OperationType = 2
	OperationType_OPERATION_TYPE_UNLOCK_BALANCE        OperationType = 3
	OperationType_OPERATION_TYPE_CREATE_TRANSFER       OperationType = 4
	OperationType_OPERATION_TYPE_TRANSFER              OperationType = 5
	OperationType_OPERATION_TYPE_MATCH_ORDER           OperationType = 6
	OperationType_OPERATION_TYPE_CREATE_ORDER_RESPONSE OperationType = 7
	OperationType_OPERATION_TYPE_RECREATE_ORDER        OperationType = 8
)

// Enum value maps for OperationType.
var (
	OperationType_name = map[int32]string{
		0: "OPERATION_TYPE_CREATE_ORDER",
		1: "OPERATION_TYPE_LOCK_BALANCE",
		2: "OPERATION_TYPE_APPROVE_CREATION",
		3: "OPERATION_TYPE_UNLOCK_BALANCE",
		4: "OPERATION_TYPE_CREATE_TRANSFER",
		5: "OPERATION_TYPE_TRANSFER",
		6: "OPERATION_TYPE_MATCH_ORDER",
		7: "OPERATION_TYPE_CREATE_ORDER_RESPONSE",
		8: "OPERATION_TYPE_RECREATE_ORDER",
	}
	OperationType_value = map[string]int32{
		"OPERATION_TYPE_CREATE_ORDER":          0,
		"OPERATION_TYPE_LOCK_BALANCE":          1,
		"OPERATION_TYPE_APPROVE_CREATION":      2,
		"OPERATION_TYPE_UNLOCK_BALANCE":        3,
		"OPERATION_TYPE_CREATE_TRANSFER":       4,
		"OPERATION_TYPE_TRANSFER":              5,
		"OPERATION_TYPE_MATCH_ORDER":           6,
		"OPERATION_TYPE_CREATE_ORDER_RESPONSE": 7,
		"OPERATION_TYPE_RECREATE_ORDER":        8,
	}
)

func (x OperationType) Enum() *OperationType {
	p := new(OperationType)
	*p = x
	return p
}

func (x OperationType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OperationType) Descriptor() protoreflect.EnumDescriptor {
	return file_ticket_proto_enumTypes[1].Descriptor()
}

func (OperationType) Type() protoreflect.EnumType {
	return &file_ticket_proto_enumTypes[1]
}

func (x OperationType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OperationType.Descriptor instead.
func (OperationType) EnumDescriptor() ([]byte, []int) {
	return file_ticket_proto_rawDescGZIP(), []int{1}
}

type Ticket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TicketId      string        `protobuf:"bytes,1,opt,name=ticket_id,json=ticketId,proto3" json:"ticket_id,omitempty"`
	State         TicketState   `protobuf:"varint,2,opt,name=state,proto3,enum=TicketState" json:"state,omitempty"`
	OperationType OperationType `protobuf:"varint,3,opt,name=operation_type,json=operationType,proto3,enum=OperationType" json:"operation_type,omitempty"`
	Data          []byte        `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Ticket) Reset() {
	*x = Ticket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ticket_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Ticket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ticket) ProtoMessage() {}

func (x *Ticket) ProtoReflect() protoreflect.Message {
	mi := &file_ticket_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ticket.ProtoReflect.Descriptor instead.
func (*Ticket) Descriptor() ([]byte, []int) {
	return file_ticket_proto_rawDescGZIP(), []int{0}
}

func (x *Ticket) GetTicketId() string {
	if x != nil {
		return x.TicketId
	}
	return ""
}

func (x *Ticket) GetState() TicketState {
	if x != nil {
		return x.State
	}
	return TicketState_TICKET_STATE_NEW
}

func (x *Ticket) GetOperationType() OperationType {
	if x != nil {
		return x.OperationType
	}
	return OperationType_OPERATION_TYPE_CREATE_ORDER
}

func (x *Ticket) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_ticket_proto protoreflect.FileDescriptor

var file_ticket_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x94,
	0x01, 0x0a, 0x06, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x69, 0x63,
	0x6b, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x69,
	0x63, 0x6b, 0x65, 0x74, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x35, 0x0a, 0x0e, 0x6f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x0d, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x2a, 0x79, 0x0a, 0x0b, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x49, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x53,
	0x54, 0x41, 0x54, 0x45, 0x5f, 0x4e, 0x45, 0x57, 0x10, 0x00, 0x12, 0x20, 0x0a, 0x1c, 0x54, 0x49,
	0x43, 0x4b, 0x45, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x57, 0x41, 0x49, 0x54, 0x5f,
	0x50, 0x52, 0x4f, 0x43, 0x45, 0x53, 0x53, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x1b, 0x0a, 0x17,
	0x54, 0x49, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x50, 0x52, 0x4f,
	0x43, 0x45, 0x53, 0x53, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x54, 0x49, 0x43,
	0x4b, 0x45, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x03,
	0x2a, 0xc7, 0x02, 0x0a, 0x0d, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x1f, 0x0a, 0x1b, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x4f, 0x52, 0x44, 0x45,
	0x52, 0x10, 0x00, 0x12, 0x1f, 0x0a, 0x1b, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e,
	0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4c, 0x4f, 0x43, 0x4b, 0x5f, 0x42, 0x41, 0x4c, 0x41, 0x4e,
	0x43, 0x45, 0x10, 0x01, 0x12, 0x23, 0x0a, 0x1f, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x41, 0x50, 0x50, 0x52, 0x4f, 0x56, 0x45, 0x5f, 0x43,
	0x52, 0x45, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x12, 0x21, 0x0a, 0x1d, 0x4f, 0x50, 0x45,
	0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x4c, 0x4f,
	0x43, 0x4b, 0x5f, 0x42, 0x41, 0x4c, 0x41, 0x4e, 0x43, 0x45, 0x10, 0x03, 0x12, 0x22, 0x0a, 0x1e,
	0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43,
	0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x45, 0x52, 0x10, 0x04,
	0x12, 0x1b, 0x0a, 0x17, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x45, 0x52, 0x10, 0x05, 0x12, 0x1e, 0x0a,
	0x1a, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x4d, 0x41, 0x54, 0x43, 0x48, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x10, 0x06, 0x12, 0x28, 0x0a,
	0x24, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x5f, 0x52, 0x45, 0x53,
	0x50, 0x4f, 0x4e, 0x53, 0x45, 0x10, 0x07, 0x12, 0x21, 0x0a, 0x1d, 0x4f, 0x50, 0x45, 0x52, 0x41,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x52, 0x45, 0x43, 0x52, 0x45, 0x41,
	0x54, 0x45, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x10, 0x08, 0x42, 0x0a, 0x5a, 0x08, 0x2f, 0x74,
	0x69, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ticket_proto_rawDescOnce sync.Once
	file_ticket_proto_rawDescData = file_ticket_proto_rawDesc
)

func file_ticket_proto_rawDescGZIP() []byte {
	file_ticket_proto_rawDescOnce.Do(func() {
		file_ticket_proto_rawDescData = protoimpl.X.CompressGZIP(file_ticket_proto_rawDescData)
	})
	return file_ticket_proto_rawDescData
}

var file_ticket_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_ticket_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ticket_proto_goTypes = []interface{}{
	(TicketState)(0),   // 0: TicketState
	(OperationType)(0), // 1: OperationType
	(*Ticket)(nil),     // 2: Ticket
}
var file_ticket_proto_depIdxs = []int32{
	0, // 0: Ticket.state:type_name -> TicketState
	1, // 1: Ticket.operation_type:type_name -> OperationType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_ticket_proto_init() }
func file_ticket_proto_init() {
	if File_ticket_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ticket_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Ticket); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ticket_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ticket_proto_goTypes,
		DependencyIndexes: file_ticket_proto_depIdxs,
		EnumInfos:         file_ticket_proto_enumTypes,
		MessageInfos:      file_ticket_proto_msgTypes,
	}.Build()
	File_ticket_proto = out.File
	file_ticket_proto_rawDesc = nil
	file_ticket_proto_goTypes = nil
	file_ticket_proto_depIdxs = nil
}
