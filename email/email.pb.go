// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: email.proto

package email // import "v2.staffjoy.com/email"

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import empty "github.com/golang/protobuf/ptypes/empty"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type EmailRequest struct {
	To                   string   `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	Subject              string   `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	HtmlBody             string   `protobuf:"bytes,3,opt,name=html_body,json=htmlBody,proto3" json:"html_body,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmailRequest) Reset()         { *m = EmailRequest{} }
func (m *EmailRequest) String() string { return proto.CompactTextString(m) }
func (*EmailRequest) ProtoMessage()    {}
func (*EmailRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_email_f6da33375fda251a, []int{0}
}
func (m *EmailRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmailRequest.Unmarshal(m, b)
}
func (m *EmailRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmailRequest.Marshal(b, m, deterministic)
}
func (dst *EmailRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmailRequest.Merge(dst, src)
}
func (m *EmailRequest) XXX_Size() int {
	return xxx_messageInfo_EmailRequest.Size(m)
}
func (m *EmailRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_EmailRequest.DiscardUnknown(m)
}

var xxx_messageInfo_EmailRequest proto.InternalMessageInfo

func (m *EmailRequest) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *EmailRequest) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *EmailRequest) GetHtmlBody() string {
	if m != nil {
		return m.HtmlBody
	}
	return ""
}

func (m *EmailRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*EmailRequest)(nil), "staffjoy.email.EmailRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// EmailServiceClient is the client API for EmailService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EmailServiceClient interface {
	Send(ctx context.Context, in *EmailRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type emailServiceClient struct {
	cc *grpc.ClientConn
}

func NewEmailServiceClient(cc *grpc.ClientConn) EmailServiceClient {
	return &emailServiceClient{cc}
}

func (c *emailServiceClient) Send(ctx context.Context, in *EmailRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/staffjoy.email.EmailService/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EmailServiceServer is the server API for EmailService service.
type EmailServiceServer interface {
	Send(context.Context, *EmailRequest) (*empty.Empty, error)
}

func RegisterEmailServiceServer(s *grpc.Server, srv EmailServiceServer) {
	s.RegisterService(&_EmailService_serviceDesc, srv)
}

func _EmailService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmailServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/staffjoy.email.EmailService/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmailServiceServer).Send(ctx, req.(*EmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _EmailService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "staffjoy.email.EmailService",
	HandlerType: (*EmailServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _EmailService_Send_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "email.proto",
}

func init() { proto.RegisterFile("email.proto", fileDescriptor_email_f6da33375fda251a) }

var fileDescriptor_email_f6da33375fda251a = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x8e, 0x41, 0x4f, 0x83, 0x40,
	0x10, 0x85, 0x2d, 0x12, 0xb5, 0xab, 0xe9, 0x61, 0x12, 0x75, 0xd3, 0x7a, 0x30, 0x3d, 0x79, 0x5a,
	0x92, 0x7a, 0xf7, 0xd0, 0xa4, 0x57, 0x0f, 0xed, 0xcd, 0x8b, 0x61, 0x61, 0x40, 0x08, 0xcb, 0x20,
	0x0c, 0x24, 0xfb, 0xef, 0x0d, 0x83, 0x44, 0x7b, 0xdb, 0xfd, 0xde, 0x7b, 0x99, 0x4f, 0xdd, 0xa2,
	0x8b, 0x8b, 0xca, 0x34, 0x2d, 0x31, 0xc1, 0xaa, 0xe3, 0x38, 0xcb, 0x4a, 0xf2, 0x46, 0xe8, 0x7a,
	0x93, 0x13, 0xe5, 0x15, 0x46, 0x92, 0xda, 0x3e, 0x8b, 0xd0, 0x35, 0xec, 0xa7, 0xf2, 0xb6, 0x50,
	0x77, 0x87, 0xb1, 0x75, 0xc4, 0xef, 0x1e, 0x3b, 0x86, 0x95, 0x0a, 0x98, 0xf4, 0xe2, 0x79, 0xf1,
	0xb2, 0x3c, 0x06, 0x4c, 0xa0, 0xd5, 0x75, 0xd7, 0xdb, 0x12, 0x13, 0xd6, 0x81, 0xc0, 0xf9, 0x0b,
	0x1b, 0xb5, 0xfc, 0x62, 0x57, 0x7d, 0x5a, 0x4a, 0xbd, 0xbe, 0x94, 0xec, 0x66, 0x04, 0x7b, 0x4a,
	0x3d, 0x80, 0x0a, 0xeb, 0xd8, 0xa1, 0x0e, 0x85, 0xcb, 0x7b, 0xf7, 0xfe, 0x7b, 0xea, 0x84, 0xed,
	0x50, 0x24, 0x08, 0x6f, 0x2a, 0x3c, 0x61, 0x9d, 0xc2, 0x93, 0x39, 0x17, 0x36, 0xff, 0x85, 0xd6,
	0x0f, 0x66, 0xd2, 0x37, 0xb3, 0xbe, 0x39, 0x8c, 0xfa, 0xdb, 0x8b, 0xfd, 0xe3, 0xc7, 0xfd, 0xb0,
	0xfb, 0xdb, 0x26, 0xe4, 0x22, 0xd9, 0xdb, 0x2b, 0xa9, 0xbe, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff,
	0x6a, 0xd3, 0xba, 0xaf, 0x16, 0x01, 0x00, 0x00,
}
