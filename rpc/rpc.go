package rpc

import (
	"reflect"
	"sync"
	"time"
)

type RpcRequest struct {
	RpcRequestData IRpcRequestData

	bLocalRequest bool
	localReply interface{}
	localParam interface{} //本地调用的参数列表
	inputArgs IRawInputArgs

	requestHandle RequestHandler
	callback *reflect.Value
	rpcProcessor IRpcProcessor
}

type RpcResponse struct {
	RpcResponseData IRpcResponseData
}

type Responder = RequestHandler

func (r *Responder) IsInvalid() bool {
	return reflect.ValueOf(*r).Pointer() == reflect.ValueOf(reqHandlerNull).Pointer()
}

//var rpcResponsePool sync.Pool
var rpcRequestPool sync.Pool
var rpcCallPool sync.Pool


type IRpcRequestData interface {
	GetSeq() uint64
	GetServiceMethod() string
	GetInParam() []byte
	IsNoReply() bool
}

type IRpcResponseData interface {
	GetSeq() uint64
	GetErr() *RpcError
	GetReply() []byte
}

type IRawInputArgs interface {
	GetRawData() []byte  //获取原始数据
	DoGc()           //处理完成,回收内存
}

type RpcHandleFinder interface {
	FindRpcHandler(serviceMethod string) IRpcHandler
}

type RequestHandler func(Returns interface{},Err RpcError)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Reply         interface{}
	Response      *RpcResponse
	Err           error
	done          chan *Call  // Strobes when call is complete.
	connId        int
	callback      *reflect.Value
	rpcHandler    IRpcHandler
	callTime      time.Time
}

func init(){
	rpcRequestPool.New = func() interface{} {
		return &RpcRequest{}
	}

	rpcCallPool.New = func() interface{} {
		return &Call{done:make(chan *Call,1)}
	}
}

func (slf *RpcRequest) Clear() *RpcRequest{
	slf.RpcRequestData = nil
	slf.localReply = nil
	slf.localParam = nil
	slf.requestHandle = nil
	slf.callback = nil
	slf.bLocalRequest = false
	slf.inputArgs = nil
	slf.rpcProcessor = nil
	return slf
}

func (rpcResponse *RpcResponse) Clear() *RpcResponse{
	rpcResponse.RpcResponseData = nil
	return rpcResponse
}

func (call *Call) Clear() *Call{
	call.Seq = 0
	call.ServiceMethod = ""
	call.Reply = nil
	call.Response = nil
	if len(call.done)>0 {
		call.done = make(chan *Call,1)
	}

	call.Err = nil
	call.connId = 0
	call.callback = nil
	call.rpcHandler = nil
	return call
}

func (call *Call) Done() *Call{
	return <-call.done
}

func MakeRpcRequest() *RpcRequest{
	return rpcRequestPool.Get().(*RpcRequest).Clear()
}

func MakeCall() *Call {
	return rpcCallPool.Get().(*Call).Clear()
}

func ReleaseRpcRequest(rpcRequest *RpcRequest){
	rpcRequestPool.Put(rpcRequest)
}

func ReleaseCall(call *Call){
	rpcCallPool.Put(call)
}

