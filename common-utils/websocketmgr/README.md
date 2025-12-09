## **WebSocket框架使用简介**       

**【1. 组件简介】**  
websocketmgr包基于开源的gorilla WebSocket框架，对基础WebSocket通信功能进一步封装，实现了客户端代理，服务端代理，连接管理，配置管理等功能模块。业务代码引用本组件后可以方便快速的实现WebSocket通讯，专注业务代码开发，无须关注具体的通信细节。 

**【2. 组件关系说明】**  
在本WebSocket组件基础的通信和业务处理过程中会使用到以下三个包：  
`websocketmgr`：负责WebSokcet连接的建立和连接状态管理。   
`limiter`: 对客户端或者服务端接收的消息数据进行限流处理(若启用限流功能)   
`modulemgr`：当前MEF业务采用基于Golang channel的业务消息路由处理模型，客户端和服务端启动时注册业务消息key与消息处理channel的映射关系，待通信端接收到消息后，根据消息的key在映射关系中找到处理消息的channel，将消息放入channel后，待channel绑定的处理函数处理完消息后返回结果，再通过WebSocket链路将处理结果返回给对端。   
   
以上三个模块关系如下所示：      
![](https://cloudmodelingapi.tools.huawei.com/cloudmodelingdrawiosvr/d/25df9cf5fcd6495e9f6e01e892b5e7cd?t=1707102362260)    

**【3. 框架使用方法】**  
1、在项目go mod文件中引入websocketmgr包  
```
require (
    codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/websocketmgr
)
```   
2、在源码中import导入websocketmgr包以及相关的包，建议遵照公司统一的包名前缀   
```
import (
    "huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/websocketmgr"
)
```   

**【4. 代码示例】**   
**服务端代码Demo:**   
1、(**必须**) 初始化代理配置，包括服务端名称，监听地址，监听端口，tls配置信息，自定义的服务端URL(若不传递则默认使用/)以及超时配置等，其中serverTlsInfo参数证书配置信息可以参照x509包的使用说明。
```
 // 1. init proxy config data （required)
proxyConfig, err := websocketmgr.InitProxyConfig(serverName, serverIp, serverPort, serverTlsInfo)
if err != nil {
	return nil, errors.New("init server proxy config failed")
}
proxyConfig.SetTimeout(readerTimeout, writeTimeout, readHeaderTimeout)
```
2、(**可选**) 根据需要设置请求频率和带宽限流器，使用指导请参照limiter包。   
```
// 2. set limiters (optional)
if err := proxyConfig.SetBandwidthLimiterCfg(1, 1); err != nil {
	return nil, err
}
if err := proxyConfig.SetRpsLimiterCfg(1.0, 1); err != nil {
	return nil, err
}
```   
3、(**必须**) 注册消息处理路由配置。   
RegModInfos函数的参数用来定义与实际业务相关的消息处理路由，其类型是modulemgr.MessageHandlerIntf接口的切片，意味着可以注册多条消息处理路由。当前由modulemgr.RegisterModuleInfo结构体实现其接口。关键字段如下:   
`MsgOpt`: MEF业务消息的option，代表动作，例如GET, POST。   
`MsgRes`: MEF业务消息的资源，即option操作的对象。   
`ModuleName`：基于Golang channel的业务消息路由处理模型中关联了处理函数的channel名称，将消息转交给相对应的channel即可完成相关的业务处理。   
MEF业务当前映射规则为MsgOpt + ":" + MsgRes作为key，ModuleName即channel名称为value，key可以保证全局唯一，将这些映射关系存储到map结构。   
**注意:** 与ModuleName关联的channel在使用前必须确保已正确初始化，当前初始化过程是由modulemgr包的Registry函数完成，具体使用方法请参照modulemgr包使用说明。     
```
// 3. register message handler map（required)
testRegInfo := []modulemgr.MessageHandlerIntf{
	&modulemgr.RegisterModuleInfo{MsgOpt: "test_opt1", MsgRes: "test_res1", ModuleName:"test_handler_module1"},
	&modulemgr.RegisterModuleInfo{MsgOpt: "test_opt2", MsgRes: "test_res2", ModuleName: "test_handler_module2"},
}
proxyConfig.RegModInfos(testRegInfo)
```   
4、(**必须**) 初始化服务端代理实例对象，使用前面步骤中初始化的配置作为服务端对象参数。   
```
// 4. init server proxy instance （required)
serverProxy := &websocketmgr.WsServerProxy{
	ProxyCfg: proxyConfig,
}
```   
5、(**必须**) 注册服务端路由，处理客户端握手请求和收发业务消息。      
WebSocket通信过程包括握手和数据收发2个阶段。此处注册的路由handler即完成这两个阶段的操作。   
握手阶段依赖于标准的HTTP协议，数据收发则依赖WebSocket协议。      
通常情况握手是一次性的，借助一次HTTP Get请求来完成，数据收发则是循环操作，除非达到特定退出条件才会终止。   
框架当前提供了2种注册路由的方式，即默认注册(`AddDefaultHandler`)和自定义注册(`AddHandler`)。   
**默认注册(AddDefaultHandler)**: 推荐方式。注册的路由URL为/,框架内部实现了处理客户端握手请求，数据校验，连接管理，限流，心跳检测和业务数据收发等功能，用户无需再实现这些操作。其中业务消息的路由处理是通过步骤4注册的路由配置来进行，通过消息的key将其发往对应的channel中，被自动进行业务逻辑处理，处理后的数据(若需要响应)被自动再次发往对端。       
**自定义注册(AddHandler)**: 由用户自己定义服务端URL和对应的处理函数，处理函数内部用户需要实现处理握手请求，连接管理，限流和业务数据收发等操作，除此之外，自定义注册方式中，步骤4的消息路由配置将不会被使用，用户需要自行维护消息和对应的handler之间的映射关系。      
```
// 5. register server handler for WebSocket handshake and message handling  (required)
serverProxy.AddDefaultHandler()
if err = serverProxy.AddHandler(demoUrl, demoHandler); err != nil {
	return nil, errors.New("add handler failed")
}
```   
6、(**必须**) 启动服务端，监听客户端握手请求。    
服务端启动的过程中会完成握手监听路由注册， HTTP服务器的启动，限流器的启动等操作。待Start完成后，服务端处于ready状态，等待客户端连接。
```
// 6. start server proxy (required), then server is ready
if err = serverProxy.Start(); err != nil {
	return nil, errors.New("server starts failed")
}
```   
7、(**可选**) 手动发送数据    
WebSocket连接建立后，可以根据需要手动调用服务端代理实例的`Send`方法发送数据。   
注意:由于服务端会同时对接多个客户端，所以服务端需要维护客户端id(通常是ip)和WebSocket连接对象的映射关系，服务端手动发送数据时通过对应的id找到连接对象，然后发送数据到客户端。   
如果步骤5采用的是默认注册方式，则无需关心客户端id和连接对象的映射关系，框架已经实现。         
```
// 7. send data manually (optional)
testMsg := getMsg()
clientId := getClientId(serverProxy)
if err := serverProxy.Send(clientId, testMsg); err != nil {
	return nil, errors.New("client send message failed")
}
```

**客户端代码Demo:**   
1、(**必须**) 初始化代理配置，包括客户端名称，监听地址，监听端口，tls配置信息，自定义的服务端URL(若不传递则默认使用/)以及超时配置等，其中clientTlsInfo参数证书配置信息可以参照x509包的使用说明。
```
 // 1. init proxy config data （required)
proxyConfig, err := websocketmgr.InitProxyConfig(clientName, serverIp, serverPort, clientTlsInfo)
if err != nil {
	return nil, errors.New("init client proxy config failed")
}
```
2、(**可选**) 根据需要设置请求频率和带宽限流器，使用指导请参照limiter包。   
```
// 2. set limiters (optional)
if err := proxyConfig.SetBandwidthLimiterCfg(1, 1); err != nil {
	return nil, err
}
if err := proxyConfig.SetRpsLimiterCfg(1.0, 1); err != nil {
	return nil, err
}
```   
3、(**必须**) 注册消息处理路由配置。   
由于WebSocket协议的全双工特性，服务端和客户端在处理业务时是对等的，均可主动向对端推送数据，所以客户端注册消息处理路由时使用的方法和注意事项与服务端一致，可以参照服务端Demo第3条说明。     
```
// 3. register message handler map（required)
testRegInfo := []modulemgr.MessageHandlerIntf{
	&modulemgr.RegisterModuleInfo{MsgOpt: "test_opt1", MsgRes: "test_res1", ModuleName:"test_handler_module1"},
	&modulemgr.RegisterModuleInfo{MsgOpt: "test_opt2", MsgRes: "test_res2", ModuleName: "test_handler_module2"},
}
proxyConfig.RegModInfos(testRegInfo)
```   
4、(**必须**) 初始化客户端代理实例对象，使用前面步骤中初始化的配置作为客户端对象参数。  
``` 
// 4. init client proxy instance (required)
clientProxy := &websocketmgr.WsClientProxy{
	ProxyCfg: proxyConfig,
}
```   
5、(**必须**) 启动客户端，尝试与服务端握手并建立WebSocket连接。    
服务端启动的过程中会完成握手监听路由注册， HTTP服务器的启动，限流器的启动等操作。待Start完成后，服务端处于ready状态，等待客户端连接。
```
// 5. start client proxy (required), then WebSocket connection is established
if err = clientProxy.Start(); err != nil {
	return nil, errors.New("client start failed")
}
```
6、(**可选**) 手动发送数据    
WebSocket连接建立后，可以根据需要手动调用客户端代理实例的`Send`方法发送数据
```
// 6. send data manually (optional)
testMsg := getMsg()
if err := clientProxy.Send(testMsg); err != nil {
	return nil, errors.New("client send message failed")
}
```