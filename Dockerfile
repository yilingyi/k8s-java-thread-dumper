# 使用多阶段构建来缓存依赖项并构建应用程序
# 阶段 1: 下载 go 依赖
FROM registry.cn-hangzhou.aliyuncs.com/yilingyi/golang:1.20.7 as go_dep
WORKDIR /source
COPY go.* ./
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go mod download

# 阶段 2: 构建应用程序
FROM registry.cn-hangzhou.aliyuncs.com/yilingyi/golang:1.20.7 as builder
WORKDIR /source
COPY --from=go_dep /go /go
COPY . .

RUN go build -o target/jdd cmd/main.go \
    && cp crawl.sh target/ \
    && chmod +x target/crawl.sh \
    && mkdir -p target/stacks

# 阶段 3: 创建最终镜像
FROM registry.cn-hangzhou.aliyuncs.com/yilingyi/ubuntu:23.04
WORKDIR /app
ENV LANG en_US.UTF-8
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone
RUN apt update && apt install -y tzdata --no-install-recommends && rm -rf /var/lib/apt/lists/*
COPY --from=builder /source/target .
COPY --from=builder /source/tools ./tools

CMD ["./jdd"]