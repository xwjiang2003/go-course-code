# 《Go 语言从入门到精通》示例源码

Gitee：https://gitee.com/jxw1111/go-course-code
GitHub：https://github.com/xwjiang2003/go-course-code

配合 B 站课程视频使用的示例代码，按集数分目录管理。所有代码都在本机 `go build` / `go vet` 验证过，标了「可运行」的能直接 `go run` 出结果。

## 目录

| 目录 | 对应视频 | 内容 | 验证方式 |
|---|---|---|---|
| `ep21-leak/` | 第21集·进阶篇：三个看不见的泄漏 | 无缓冲 channel、忘 cancel 的请求、worker 不退休，各含泄漏版与修复版 | `go build` + `go vet` + 4 个 `go run` + `go test ./p16` |
| `ep36-grpc/` | 第36集·进阶篇：gRPC 生产化 | 秒杀下单调用链：预算、取消、服务端流、连接复用 | `go build` + `go vet` + 4 个 `go run` |
| `ep38-cache/` | 第38集·进阶篇：缓存与消息 | 缓存击穿、版本 CAS 回填、Kafka 重复消费 | `go build` + `go vet` |
| `ep41-shortlink/` | 第41集·实战：短链服务 | 短链生成、302 跳转、缓存、限流、单元测试 | `go build` + `go vet` + `go test` + 启动服务 |

## 环境要求

- Go 1.23 及以上（`ep38-cache` 与 `ep41-shortlink` 在 Go 1.21 / 1.22 也能编译）
- 只有 `ep38-cache` 运行示例时需要本地 Redis / Kafka；`ep36-grpc`、`ep21-leak` 与 `ep41-shortlink` 不依赖任何外部服务

## ep21-leak：三个看不见的泄漏

三个案例都分 `leaky/` 与 `fixed/` 两个可独立运行的程序，另有一个合并版和一个单测：

```bash
cd ep21-leak

go run ./p06/leaky   # 无缓冲 channel 发送没人收：2 → 302，精确 +300
go run ./p06/fixed   # 缓冲 + select：2 → 2，全部自己退出
go run ./p09/leaky   # 忘 cancel 的请求：50 次 → +251（一次漏 5 个）
go run ./p09/fixed   # 四条保命线：2 → 2
go run ./p12/leaky   # worker range 不 close：20 轮 × 10 → +200
go run ./p12/fixed   # close + ctx.Done 两条出口：1 → 1
go run ./p14         # 三案例合一，跑完直接打印 pprof 分组表（total 757）

go test ./p16 -run TestFixedDispatch -v   # PASS
go test ./p16 -run TestLeakyDispatch -v   # FAIL：测试函数 PASS，goleak 在最后报残留栈
```

| 目录 | 视频里的位置 | 讲的是什么 |
|---|---|---|
| `p06/` | P06–P08「发送方永远卡在那一行」 | 无缓冲 channel + 接收方提前 return；修复是缓冲 1 + `select` 听 `ctx.Done()` |
| `p09/` | P09–P11「它看起来用了 context」 | ctx 没超时、没 `defer cancel`、Client 没 Timeout；一次请求漏 5 个 |
| `p12/` | P12–P13「range 的结束条件是 close」 | worker 热更新后旧的不退休；修复是两条退出路径 + 明确的 stop |
| `p14/` | P14「debug=1 给你分组表」 | 三案例合一，起 pprof 端口抓 `?debug=1`，只留组头与最内层一帧 |
| `p16/` | P16「单测兜底」 | `goleak.VerifyTestMain` 放进 `TestMain`，残留 goroutine 整体判失败 |

`p06` / `p09` / `p12` 都支持 `-wait` 参数：跑完挂住不退出，方便自己 `curl` pprof 端点看栈。

## ep36-grpc：一次秒杀，四个服务

四个可独立运行的演示，全部只在本机起服务，不依赖数据库：

```bash
cd ep36-grpc

go run ./p05   # 无预算 vs 400ms 预算：802ms vs 400ms，后者返回 DeadlineExceeded
go run ./p08   # 客户端取消 → 服务端收到 Canceled，回滚不扣减
go run ./p11   # 服务端流推 4 条排队位置，客户端断开后服务端立刻停手
go run ./p13   # 200 次调用共用 1 条 ClientConn，打印平均耗时
```

| 目录 | 视频里的位置 | 讲的是什么 |
|---|---|---|
| `p05/` | P07「网关：所有等待都要有上限」 | `context.WithTimeout` 逐层递减预算，超时快速失败 |
| `p08/` | P10「库存：先听取消，再干活」 | `select` 监听 `ctx.Done()`，取消时返回标准 `Canceled` |
| `p11/` | P13「服务端流：两个必须处理的边界」 | `Send` 失败立刻退出；客户端 `io.EOF` 是正常结束 |
| `p13/` | P15「一条 ClientConn 走天下」 | `grpc.NewClient` + resolver + `round_robin` + keepalive |

`proto/` 是接口定义，`gen/` 是用 `buf` + `protoc-gen-go` / `protoc-gen-go-grpc` 生成的真实桩代码（没有手写）。需要重新生成时（先 `go install` 两个插件并装好 buf）：

```bash
cd ep36-grpc
buf generate
```

## ep38-cache：缓存与消息队列

三个包对应视频里的三段关键代码，编译检查即可：

```bash
cd ep38-cache
go build ./...
go vet ./...
```

| 目录 | 讲的是什么 |
|---|---|
| `p05/` | 缓存击穿：singleflight 合并回源 + 空值缓存 |
| `p09/` | 版本号 CAS 回填：旧值不覆盖新值（Lua 脚本保证原子） |
| `p12/` | Kafka 消费与幂等：重复消息怎么识别、怎么丢弃 |

这三个包依赖 Redis / Kafka 客户端，作为可复用的代码片段提供，不直接 `go run`。

## ep41-shortlink：短链服务

零外部依赖，内存存储，直接可跑：

```bash
cd ep41-shortlink
go test ./...          # base62 / handler 测试通过
go run ./cmd/server    # 监听 :8080
```

另开一个终端：

```bash
curl -X POST http://127.0.0.1:8080/shorten \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://example.com/go/grpc\"}"
# {"code":"1","url":"https://example.com/go/grpc"}

curl -i http://127.0.0.1:8080/1     # 302 → https://example.com/go/grpc
```

想换成 MySQL 存储，把 `cmd/server/main.go` 里的 `store.NewMemory()` 换成 `store.NewMySQL(db)`；`docker-compose.yml` 里有一份 MySQL 8 的编排可以直接用。

| 目录 | 内容 |
|---|---|
| `internal/base62/` | 62 进制编解码 |
| `internal/cache/` | 本地缓存 + 过期 |
| `internal/store/` | 存储接口、内存实现、MySQL 实现 |
| `internal/service/` | 短链生成与解析（含基准测试） |
| `internal/handler/` | HTTP 接口：`POST /shorten`、`GET /{code}` |
| `internal/middleware/` | 日志、panic 恢复 |

## 说明

- 代码用于教学，数据都是示例，不代表任何真实系统的指标。
- 例子刻意保持短小，重点在「看得懂、跑得起来」，生产使用请补上鉴权、限流、监控与错误处理。
