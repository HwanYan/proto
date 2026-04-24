#!/bin/bash -e

# Proto 文件编译脚本
# 生成：pb.go（消息）、_grpc.pb.go（gRPC 服务端/客户端）、pb.gw.go（HTTP 网关）、swagger.json（API 文档）

# 脚本所在目录（helloworld/）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# 项目根目录（mildlab-proto/）
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

GOBIN="$(go env GOPATH)/bin"
# google/api/*.proto 的标准来源：grpc-gateway v1 自带的 third_party/googleapis
GOOGLEAPIS="$(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis"

echo "🚀 开始编译所有 .proto 文件..."
echo "📁 项目根目录: $ROOT_DIR"

# 查找当前目录下所有 .proto 文件（排除 google/api 等第三方文件）
proto_files=$(find "$SCRIPT_DIR" -name "*.proto" -type f)

if [ -z "$proto_files" ]; then
    echo "❌ 未找到任何 .proto 文件"
    exit 1
fi

echo "📋 找到以下 .proto 文件:"
echo "$proto_files" | while read file; do
    echo "  - $file"
done

echo ""
echo "🔧 开始编译..."

for proto_file in $proto_files; do
    echo ""
    echo "🔄 编译: $proto_file"

    # 1. 生成消息代码（*.pb.go）
    protoc \
        -I "$SCRIPT_DIR" \
        -I "$GOOGLEAPIS" \
        --go_out="$SCRIPT_DIR" \
        --go_opt=paths=source_relative \
        "$proto_file"
    echo "✅ 消息代码生成成功（*.pb.go）"

    # 2. 生成 gRPC 服务端/客户端代码（*_grpc.pb.go）
    if [ -f "$GOBIN/protoc-gen-go-grpc" ]; then
        protoc \
            -I "$SCRIPT_DIR" \
            -I "$GOOGLEAPIS" \
            --plugin=protoc-gen-go-grpc="$GOBIN/protoc-gen-go-grpc" \
            --go-grpc_out="$SCRIPT_DIR" \
            --go-grpc_opt=paths=source_relative \
            "$proto_file"
        echo "✅ gRPC 服务端/客户端代码生成成功（*_grpc.pb.go）"
    else
        echo "⚠️  跳过 gRPC 代码生成（protoc-gen-go-grpc 未安装）"
        echo "   安装命令: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    fi

    # 3. 生成 HTTP 网关代码（*.pb.gw.go）
    if [ -f "$GOBIN/protoc-gen-grpc-gateway" ]; then
        protoc \
            -I "$SCRIPT_DIR" \
            -I "$GOOGLEAPIS" \
            --plugin=protoc-gen-grpc-gateway="$GOBIN/protoc-gen-grpc-gateway" \
            --grpc-gateway_out="$SCRIPT_DIR" \
            --grpc-gateway_opt=paths=source_relative \
            "$proto_file"
        echo "✅ HTTP 网关代码生成成功（*.pb.gw.go）"
    else
        echo "⚠️  跳过 HTTP 网关代码生成（protoc-gen-grpc-gateway 未安装）"
        echo "   安装命令: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"
    fi

    # 4. 生成 OpenAPI 文档（*.swagger.json）
    if [ -f "$GOBIN/protoc-gen-openapiv2" ]; then
        protoc \
            -I "$SCRIPT_DIR" \
            -I "$GOOGLEAPIS" \
            --plugin=protoc-gen-openapiv2="$GOBIN/protoc-gen-openapiv2" \
            --openapiv2_out="$SCRIPT_DIR" \
            --openapiv2_opt=logtostderr=true \
            "$proto_file"
        echo "✅ OpenAPI 文档生成成功（*.swagger.json）"
    else
        echo "⚠️  跳过 OpenAPI 文档生成（protoc-gen-openapiv2 未安装）"
        echo "   安装命令: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"
    fi
done

echo ""
echo "🎉 所有 .proto 文件编译完成！"
echo ""
echo "📂 生成的文件:"
find "$SCRIPT_DIR" -name "*.pb.go" -o -name "*_grpc.pb.go" -o -name "*.pb.gw.go" -o -name "*.swagger.json" | sort | while read file; do
    echo "  - ${file#$ROOT_DIR/}"
done

echo ""
echo "📚 文件说明:"
echo "  *.pb.go        - protobuf 消息定义（序列化/反序列化）"
echo "  *_grpc.pb.go   - gRPC 服务端接口 & 客户端 Stub"
echo "  *.pb.gw.go     - HTTP/REST 网关代码（grpc-gateway）"
echo "  *.swagger.json - OpenAPI 2.0 接口文档"