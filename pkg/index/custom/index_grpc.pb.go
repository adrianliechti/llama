// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: index.proto

package custom

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Index_List_FullMethodName   = "/index.index/List"
	Index_Delete_FullMethodName = "/index.index/Delete"
	Index_Index_FullMethodName  = "/index.index/Index"
	Index_Query_FullMethodName  = "/index.index/Query"
)

// IndexClient is the client API for Index service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IndexClient interface {
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*Documents, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Index(ctx context.Context, in *IndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*Results, error)
}

type indexClient struct {
	cc grpc.ClientConnInterface
}

func NewIndexClient(cc grpc.ClientConnInterface) IndexClient {
	return &indexClient{cc}
}

func (c *indexClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*Documents, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Documents)
	err := c.cc.Invoke(ctx, Index_List_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Index_Delete_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) Index(ctx context.Context, in *IndexRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Index_Index_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*Results, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Results)
	err := c.cc.Invoke(ctx, Index_Query_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IndexServer is the server API for Index service.
// All implementations must embed UnimplementedIndexServer
// for forward compatibility.
type IndexServer interface {
	List(context.Context, *ListRequest) (*Documents, error)
	Delete(context.Context, *DeleteRequest) (*emptypb.Empty, error)
	Index(context.Context, *IndexRequest) (*emptypb.Empty, error)
	Query(context.Context, *QueryRequest) (*Results, error)
	mustEmbedUnimplementedIndexServer()
}

// UnimplementedIndexServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedIndexServer struct{}

func (UnimplementedIndexServer) List(context.Context, *ListRequest) (*Documents, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedIndexServer) Delete(context.Context, *DeleteRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedIndexServer) Index(context.Context, *IndexRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Index not implemented")
}
func (UnimplementedIndexServer) Query(context.Context, *QueryRequest) (*Results, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}
func (UnimplementedIndexServer) mustEmbedUnimplementedIndexServer() {}
func (UnimplementedIndexServer) testEmbeddedByValue()               {}

// UnsafeIndexServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IndexServer will
// result in compilation errors.
type UnsafeIndexServer interface {
	mustEmbedUnimplementedIndexServer()
}

func RegisterIndexServer(s grpc.ServiceRegistrar, srv IndexServer) {
	// If the following call pancis, it indicates UnimplementedIndexServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Index_ServiceDesc, srv)
}

func _Index_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Index_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Index_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_Index_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IndexRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).Index(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Index_Index_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).Index(ctx, req.(*IndexRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Index_Query_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Index_ServiceDesc is the grpc.ServiceDesc for Index service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Index_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "index.index",
	HandlerType: (*IndexServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Index_List_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Index_Delete_Handler,
		},
		{
			MethodName: "Index",
			Handler:    _Index_Index_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _Index_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "index.proto",
}
