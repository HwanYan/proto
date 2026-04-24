# RPC 代码生成项目

这是一个使用 Protocol Buffers 和 gRPC 网关生成 RPC 代码的示例项目。

## 项目结构

```
├── helloworld.proto          # Proto 定义文件
├── proto.sh                  # 编译脚本
├── install-trpc-plugin.sh    # 插件安装脚本
├── example/                  # 使用示例
│   └── main.go              # 示例代码
└── README.md                 # 说明文档
```

## 生成的代码类型

运行 `./proto.sh` 后会生成以下文件：

1. **helloworld.pb.go** - 标准 protobuf 消息定义
2. **helloworld.pb.gw.go** - gRPC 网关代码（HTTP REST API）
3. **helloworld.swagger.json** - OpenAPI 文档

## 快速开始

### 1. 编译 Proto 文件

```bash
./proto.sh
```

### 2. 运行示例

```bash
cd example
go run main.go
```

### 3. 测试 API

#### gRPC 客户端测试
```bash
# 运行示例程序会自动测试 gRPC 客户端
```

#### HTTP REST API 测试
```bash
curl -X POST http://localhost:8080/v1/hello \
  -H "Content-Type: application/json" \
  -d '{"msg": "World"}'
```

#### 查看 API 文档
打开 `helloworld.swagger.json` 文件查看完整的 API 规范。

## API 接口

### gRPC 服务
- **服务名**: Greeter
- **方法**: SayHello
- **端口**: 50051

### HTTP REST API
- **URL**: POST /v1/hello
- **Content-Type**: application/json
- **请求体**: `{"msg": "string"}`
- **端口**: 8080

## 技术栈

- **Protocol Buffers**: 接口定义语言
- **gRPC**: 高性能 RPC 框架
- **gRPC Gateway**: gRPC 到 HTTP REST 的转换
- **OpenAPI**: API 文档规范

## 优势

1. **高性能**: gRPC 使用 HTTP/2 和 Protocol Buffers
2. **多语言支持**: 支持多种编程语言
3. **REST 兼容**: 通过网关提供 HTTP REST API
4. **自动文档**: 生成 OpenAPI 规范
5. **强类型**: 编译时类型检查

## 扩展使用

### 添加新的 RPC 方法

在 `helloworld.proto` 中添加新的方法：

```protobuf
service Greeter {
  rpc SayHello (HelloRequest) returns (HelloResponse) {
    option (google.api.http) = {
      post: "/v1/hello"
      body: "*"
    };
  }
  
  rpc SayGoodbye (GoodbyeRequest) returns (GoodbyeResponse) {
    option (google.api.http) = {
      post: "/v1/goodbye"
      body: "*"
    };
  }
}
```

### 自定义网关配置

在 `proto.sh` 中可以添加更多编译选项：

```bash
# 生成更详细的 OpenAPI 文档
protoc --openapiv2_out=. --openapiv2_opt=logtostderr=true,openapi_configuration=config.yaml
```

## 故障排除

### 插件未找到
如果提示插件未安装，运行：

```bash
./install-trpc-plugin.sh
```

### 导入错误
确保生成的 `.pb.go` 文件在正确的包路径下，与 `go_package` 选项一致。