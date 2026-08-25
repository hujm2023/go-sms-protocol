# CMPP 2.0 Server 示例

本目录提供一个基于 `nioserver` 的、可以直接运行的 CMPP 2.0 接入服务。它不是只演示 API 的 `_test.go`，而是展示一个完整服务需要负责的协议处理和进程生命周期：

- 按 CMPP 2.0 规范计算 MD5 鉴权摘要，并校验 `CMPP_CONNECT` 时间戳；
- 在连接上保存鉴权状态，拒绝成功鉴权前发送的业务 PDU；
- 处理 `CMPP_SUBMIT`，返回 `CMPP_SUBMIT_RESP`，同时只记录不含敏感信息的提交元数据；
- 当 `RegisteredDelivery == 1` 时，通过有界内存队列在 100～500 ms 后推送成功的 `CMPP_DELIVER` 状态报告；
- 响应客户端发起的 `CMPP_ACTIVE_TEST`，也由服务端主动探活；连续三次未响应时关闭连接；
- 处理 `CMPP_TERMINATE`、进程信号、写超时和优雅退出。

## 运行方式

账号和密码从环境变量读取，避免凭据出现在 shell history 或进程参数列表中：

```bash
export CMPP20_ACCOUNT='900001'
export CMPP20_PASSWORD='replace-with-a-secret'

go run ./nioserver/examples/cmpp20server
```

服务默认监听 `127.0.0.1:7890`。执行 `go run ./nioserver/examples/cmpp20server -help` 可以查看全部配置，例如：

```bash
go run ./nioserver/examples/cmpp20server \
  -listen='127.0.0.1:7890' \
  -gateway-id=1 \
  -heartbeat-interval=30s \
  -receipt-min-delay=100ms \
  -receipt-max-delay=500ms
```

`CMPP_CONNECT` 的时间戳不包含年份和时区，示例按进程本地时区解释。服务端和客户端约定了特定网关时区时，应通过 `TZ` 设置进程时区。

## 提交与回执流程

服务接受 `CMPP_SUBMIT` 后，会生成网关消息 ID 并通过 `CMPP_SUBMIT_RESP` 返回。日志只记录消息 ID、序列号、账号、业务类型、接收号码数量、内容字节数以及是否请求回执；不会记录密码、短信正文或完整接收号码。

对于每个请求回执的接收号码，worker 会在配置的随机延迟后发送 `CMPP_DELIVER`，其中：

- `RegisteredDeliver = 1`；
- `Stat = DELIVRD`；
- 外层 `PduDeliver.MsgID` 是本次推送的新 ID，客户端需要在 `CMPP_DELIVER_RESP` 中回显；
- 回执正文中的 `MsgID` 是原 `CMPP_SUBMIT_RESP.MsgID`，用于关联原始提交；
- 回执正文同时携带原接收号码和提交序列号。

客户端应通过 `CMPP_DELIVER_RESP` 确认收到回执，服务端会记录该确认的元数据。

## 为什么采用这些设计

- 回执任务使用固定 worker 数量和有界队列，避免每条短信创建 goroutine，也避免高流量下无限占用内存。
- 同一个 `CMPP_SUBMIT` 的所有回执任务要么整批入队，要么返回流控错误，避免客户端收到成功响应后只有部分号码获得回执。
- 回执和主动探活使用独立的服务级 context，因为单次请求的 context 在 handler 返回后可能被取消。
- 服务端只使用一个定时器扫描已鉴权连接，避免为每条连接单独创建 ticker。
- 关闭流程先停止后台任务，再关闭监听与连接，最后在统一截止时间内等待 goroutine 退出。

## 能力边界

这是一个可执行的集成示例，不是可直接部署到生产环境的完整短信网关。它有意省略了以下能力：

- 只支持由环境变量配置的一组账号和密码；
- 时间窗口可以拒绝过期鉴权，但窗口内没有持久化的重放保护；
- 已受理提交和待发送回执只保存在内存中，进程退出后会丢失；
- 没有实现长短信重组、消息去重、持久化、计费、路由、重试、指标和审计存储；
- 消息 ID 由单进程内存序号生成，多实例部署时没有跨实例防冲突能力；
- CMPP 原始 TCP 流量没有加密，真实部署应使用可信网络或加密隧道；
- 优雅退出会取消仍在队列中等待的模拟回执。

将这里的回执流程用于真实业务前，应先接入持久化、幂等消息存储和可靠队列。
