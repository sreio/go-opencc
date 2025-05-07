# syntax=docker/dockerfile:1

# 第一阶段：构建应用程序
FROM golang:1.24 AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-opencc main.go

# 第二阶段：创建最小化运行时镜像
FROM alpine:latest

# 安装必要的证书
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/go-opencc .

# 暴露应用程序端口
EXPOSE 8581

# 设置容器启动命令
CMD ["./go-opencc"]
