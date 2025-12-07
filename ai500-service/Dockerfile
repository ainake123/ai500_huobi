# 第一阶段：编译
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖配置
COPY go.mod ./
# 下载依赖 (如果有 external lib)
# RUN go mod download

# 复制源码
COPY . .

# 编译 (CGO_ENABLED=0 确保生成静态二进制文件)
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# 第二阶段：运行
FROM alpine:latest

# 安装 CA 证书，否则无法请求 HTTPS (火币接口)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从编译阶段复制二进制文件
COPY --from=builder /app/server .

# 暴露端口
EXPOSE 2234

# 启动命令
CMD ["./server"]
