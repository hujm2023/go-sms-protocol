# SGIP 1.2 网关与 SP 示例

本目录提供基于 `nioserver.BaseServer` 的 SMG 模拟网关；[`../sgip12client`](../sgip12client) 是配套 SP 应用，既能提交短信，也会在同一进程中监听网关的回执连接。

## 为什么需要两个连接

SGIP 的命令由发起连接的一方向接收连接的一方发送，双方互为客户端和服务器：

1. SP 连接网关：`BIND(type=1) → SUBMIT → SUBMIT_RESP → UNBIND`。
2. 网关另建连接到 SP 的监听端口：`BIND(type=2) → REPORT → REPORT_RESP → UNBIND`。

本示例故意在提交完成后关闭第一条连接，再等待第二条连接完成回执。两条连接可以由同一个 SP 进程处理，SP 需要额外的监听端口；部署时该端口必须可由网关访问。

协议依据见 [SGIP 文档](../../../doc/中国联通_SGIP1.2.pdf) §3.4、§3.7.2 和 §4.2.3.5。SGIP 没有 CMPP 的 `ACTIVE_TEST` 或 SMPP 的 `enquire_link` 命令；示例使用读超时、空闲超时和 `UNBIND` 管理连接。

## 本地运行

在仓库根目录的两个终端配置相同的四个环境变量。提交连接使用 SP 账号，回连使用网关账号；两组凭据分别鉴权，每个字段最多 16 字节：

```bash
export SGIP12_ACCOUNT='sp-demo'
export SGIP12_PASSWORD='sp-secret'
export SGIP12_CALLBACK_ACCOUNT='gw-demo'
export SGIP12_CALLBACK_PASSWORD='gw-secret'

# 终端一：网关监听 8801，并将回执发往 SP 的 8802
go run ./nioserver/examples/sgip12server \
  -listen=127.0.0.1:8801 \
  -sp-address=127.0.0.1:8802
```

```bash
export SGIP12_ACCOUNT='sp-demo'
export SGIP12_PASSWORD='sp-secret'
export SGIP12_CALLBACK_ACCOUNT='gw-demo'
export SGIP12_CALLBACK_PASSWORD='gw-secret'

# 终端二：SP 先监听 8802，再连接网关提交短信
go run ./nioserver/examples/sgip12client \
  -server=127.0.0.1:8801 \
  -listen=127.0.0.1:8802
```

客户端验证回执并确认网关退出后，以退出码 0 结束；输出中的提交 ID 随运行变化：

```text
SP callback listener ready address=127.0.0.1:8802
submission bind accepted type=1
submit accepted submit_id=3000000001:1020123456:2
submission unbind acknowledged; awaiting a separate callback connection
callback bind accepted type=2
report received submit_id=3000000001:1020123456:2 state=0 acknowledged
callback unbind acknowledged
```

示例不会向真实手机发送短信。`-sp-address` 是网关侧配置的单个 SP 地址，不从 `SUBMIT` 中推断，也不会自动发现 SP 的监听端口。跨机器运行时应分别配置网关地址、SP 的监听地址和网关可访问的 SP 地址。

## 回执和序列号

SGIP 的序列号是 **三个 32 位整数组成的 12 字节值**：节点编号、`MMDDhhmmss` 时间戳、流水号。响应应完整回显请求头的三个分量。

`SUBMIT_RESP` 没有另一个由网关分配的消息 ID。关联回执时，应比较 `REPORT.SubmitSequence` 与原 `SUBMIT.Header.Sequence` 的全部 12 字节。`REPORT.Header.Sequence` 是网关为这条回执新分配的序列号，SP 的 `REPORT_RESP` 应回显该序列号。只比较 `GetSequenceID()` 返回的最后四字节会混淆不同节点或不同时刻的提交。

网关支持每次提交 1～100 个号码，号码需包含 `86` 国别码；本例将正文限制为 140 字节。`ReportFlag=1` 时，为每个号码模拟 `ReportType=0`、`State=0`、`ErrorCode=0` 的成功回执；0 只在最终失败时报告，2 不请求报告，因此两者在本例的成功模拟中都不推送。包月计费专用模式 3 不在示例范围内，返回参数错误。

固定 worker 使用有界队列，**一个队列任务包含一次提交的全部号码**；容量不足时整批拒绝并返回节点忙（11）。默认延迟 100～500 ms 后，worker 独立拨号、鉴权，依次发送各号码回执并校验确认，最后完成回连侧的 `UNBIND`。SP 会拒绝错误凭据、错误连接类型、错误提交序列号和不符合成功模拟的回执。

## 配置、验证与边界

```bash
go run ./nioserver/examples/sgip12server -help
go run ./nioserver/examples/sgip12client -help
go test ./nioserver/examples/sgip12...
go test -race -tags=integration ./nioserver/examples/sgip12server
```

网关可配置 `-node-id`、`-sp-address`、`-callback-timeout`、回执 worker/队列/延迟，以及传输和关闭超时。SP 客户端提供 `-node-id`、`-server`、`-listen`、`-source`、`-destination`、`-message` 和完整交互超时 `-timeout`（默认 30 秒）。默认网关节点编号为 `2000000001`，SP 为 `3000000001`，均为演示值，真实对接应使用分配的节点编号。

日志不输出密码、正文或完整接收号码。SIGINT/SIGTERM 触发有截止时间的退出，网关会取消排队任务及正在等待的回连。

这是单账号、单 SP 地址的内存模拟，不提供持久化、失败重试、消息去重、跨实例序列号分配、真实路由/计费或长短信重组。回连失败或回执未获确认时只记录错误，不保证投递，也不补发；进程退出会丢失未完成任务。回连鉴权使用原始 SGIP 协议字段，真实部署应采用受信任的传输环境。
