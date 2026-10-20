# SMGP 3.0 服务端与客户端示例

本目录沿用 `cmpp20server` 的组织方式，提供基于 `nioserver.BaseServer` 的可运行 SMGP 服务端，配套客户端位于 [`../smgp30client`](../smgp30client)。服务端实现登录鉴权、短信提交、模拟成功回执、双向探活和优雅关闭。

## 回执是否需要单独的监听服务

**SGIP 需要 SP 接收网关的反向连接；SMGP 可以在同一条 TCP 长连接上提交短信、接收回执。**

- SGIP：SP 主动连接 SMG，完成 `BIND → SUBMIT → SUBMIT_RESP`；回执由 SMG 另外连接 SP 的监听地址，完成 `BIND → REPORT → REPORT_RESP`。因此 SP 需要提供接收连接的端口，但可以由同一个应用进程承担，不要求拆成独立进程。依据：[SGIP 协议 §3.7.2、§3.7.2.5～6](../../../doc/中国联通_SGIP1.2.pdf)（PDF 第 14、16～17 页）。
- SMGP：客户端使用 `LoginMode=2`（收发模式）登录后，服务端直接在该连接上发送 `DELIVER` 状态报告，客户端返回 `DELIVER_RESP`，无需 SP 另开监听端口。依据：[SMGP 3.0.3 §4.2.1、§6.2.3](../../../doc/中国电信_SMGP3.0.3.pdf)（PDF 第 7、29 页）。

本示例将回执发送给原提交连接。它接受三种登录模式；接收模式（1）不能提交，发送模式（0）只能提交不要求回执的短信，要求回执的提交必须使用收发模式（2）。不支持按账号将回执转发至另一条接收连接。

## 本地运行

在仓库根目录的两个终端中配置相同的账号和密码。账号最多 8 字节，密码最多 15 字节：

```bash
export SMGP30_ACCOUNT='900001'
export SMGP30_PASSWORD='test-secret'

# 终端一：默认监听 127.0.0.1:8890
go run ./nioserver/examples/smgp30server
```

```bash
export SMGP30_ACCOUNT='900001'
export SMGP30_PASSWORD='test-secret'

# 终端二：登录、提交一条模拟短信、确认回执、验证心跳、退出
go run ./nioserver/examples/smgp30client
```

客户端成功运行后退出码为 0，输出如下（消息 ID 随运行变化，回执与心跳的先后顺序可能不同）：

```text
login accepted login_mode=2
submit accepted msg_id=...
heartbeat acknowledged
receipt received msg_id=... status=DELIVRD acknowledged
exit acknowledged
```

示例不会调用真实运营商或向手机投递短信。客户端使用 ASCII 内容，最多 140 字节；可通过 `-message`、`-source`、`-destination` 调整演示提交。服务端可通过 `-listen` 调整地址，客户端使用 `-server` 指向相同地址。

```bash
go run ./nioserver/examples/smgp30server -help
go run ./nioserver/examples/smgp30client -help
```

服务端主要配置包括 `-gateway-id`（SMGW 的六位 BCD 代码，默认 `010061`）、`-auth-skew`、`-receipt-workers`、`-receipt-queue`、`-receipt-min-delay`、`-receipt-max-delay`、`-heartbeat-interval`、`-heartbeat-misses` 和传输、关闭超时。`LOGIN.Timestamp` 没有年份或时区，两端应约定相同的时区；可通过进程的 `TZ` 环境变量设置。

## 协议处理

- 登录：校验账号、版本、模式、时间窗口及 16 字节二进制 MD5 摘要，成功响应使用四字节网络序 `Status` 计算服务端摘要；失败响应先写出再关闭连接。结果码遵循 SMGP 3.0.3 §6.2.6，例如认证错误为 21，版本过高为 22。
- 拆包：按大端总长度提取完整帧，接受 TCP 半包、粘包，在读取完整帧前限制长度为 12 字节～64 KiB，并拒绝与 PDU 结构长度不一致的帧。
- 提交：支持 MT 消息，每次最多 100 个号码、140 字节内容；返回 `SUBMIT_RESP`。要求回执时，一次提交的任务整批进入有界队列，容量不足返回系统忙（1）。
- 回执：固定 worker 在默认 100～500 ms 延迟后发送 `IsReport=1`、`MsgFormat=0` 的 `DELIVER`。正文采用 `DeliveryReceipt` 编码，其中 `id` 是原 `SUBMIT_RESP.MsgID` 对应的 **10 字节二进制值**，状态为 `DELIVRD`、错误为 `000`；外层 `DELIVER.MsgID` 是本次推送的新 ID。
- 确认：客户端在 `DELIVER_RESP` 中回显外层消息 ID 和流水号。`GenEmptyResponse` 只填写流水号，调用方需要显式设置 `MsgID`。
- 探活与关闭：响应客户端心跳，服务端也主动探活，连续三次未响应则关闭连接；`ACTIVE_TEST_RESP` 按协议仅包含 12 字节消息头。收到 `EXIT` 后先写出 `EXIT_RESP` 再关闭。进程收到 SIGINT/SIGTERM 时取消后台工作并在统一截止时间内关闭服务。

日志只记录消息 ID、流水号、账号、号码数量、内容长度等元数据，不输出密码、鉴权摘要、短信正文或完整号码。

SMGP 3.0.3 §6.2.2 和 §6.2.7 强制使用 MD5 计算线上鉴权摘要，替换算法会破坏互通。这里的计算不用于存储密码，也不应被复制为密码存储方案；实际接入仍需要可信网络或加密隧道。CodeQL 工作流只对客户端验签、服务端两处鉴权摘要的三个明确标注调用应用协议例外，其它位置和规则继续报告。

## 验证与边界

```bash
go test ./nioserver/examples/smgp30... ./smgp/...
go test -race -tags=integration ./nioserver/examples/smgp30server
```

集成测试使用本机临时端口，覆盖真实 TCP 上的登录、提交、回执与确认、双向心跳、退出，以及鉴权前提交、错误密码和心跳超时。

这是单账号、单进程的集成示例。提交和回执只保存在内存中，关闭时取消未完成回执；不提供持久化、回执重试与去重、滑动窗口、长短信重组、真实计费和路由，也不跟踪 `DELIVER_RESP` 的超时或重发。消息 ID 的序号每百万次循环，不能保证高流量或多实例部署时全局唯一。时间窗口校验不提供窗口内的重放保护。用于真实业务前，需要按运营商要求补齐这些能力。
