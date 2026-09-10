# ykt-admin-go — 运维管理后台 API
#
# 纯 Go 实现（database/sql + go-sql-driver/mysql），无 CGO 依赖，
# 因此可以完全静态编译，运行镜像用 alpine 即可。

FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN GOPROXY=https://goproxy.cn,direct go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/admin \
      ./cmd/admin

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /out/admin /app/admin
# 内置默认配置作为兜底；生产环境挂载覆盖整个 configs 目录。
COPY configs/ /app/configs/

# main.go 用 os.ReadFile("configs/config.yaml") 读相对路径，
# 工作目录必须是 /app。
EXPOSE 8090

ENTRYPOINT ["/app/admin"]
