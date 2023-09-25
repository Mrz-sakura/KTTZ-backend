# 使用CentOS作为基础镜像
FROM centos:8-slim

# 安装必要的工具
RUN yum update -y && yum install -y git wget

# 安装Golang
ENV GO_VERSION 1.20
RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin

# 设置工作目录
WORKDIR /app

# 下载Go模块依赖
RUN go mod download

# 构建Golang应用
RUN go build -o KTTZ-backend .

# 运行应用
CMD ["./KTTZ-backend"]
