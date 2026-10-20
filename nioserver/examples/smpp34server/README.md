# SMPP 3.4 服务端与客户端示例

本目录提供基于 `nioserver.BaseServer` 的可运行 SMSC 模拟服务，配套 ESME 客户端位于 [`../smpp34client`](../smpp34client)。组织方式与 CMPP、SMGP 示例一致，包含绑定鉴权、提交、模拟回执、双向心跳和优雅关闭。

## 运行

在仓库根目录的两个终端配置同一组环境变量。`system_id` 最多 15 字节，密码最多 8 字节，均使用 ASCII；字段上限已扣除协议要求的结尾 NUL：

```bash
export SMPP34_ACCOUNT='demo'
export SMPP34_PASSWORD='secret34'

# 终端一：默认监听 127.0.0.1:2775
go run ./nioserver/examples/smpp34server
```

```bash
export SMPP34_ACCOUNT='demo'
export SMPP34_PASSWORD='secret34'

# 终端二
go run ./nioserver/examples/smpp34client
```

客户端使用 `bind_transceiver`，验证以下完整流程后以退出码 0 结束：

```text
bind accepted mode=transceiver
submit accepted msg_id=1
heartbeat acknowledged
receipt received msg_id=1 status=DELIVRD acknowledged
unbind acknowledged
```

消息 ID 随进程内提交次数增长，心跳和回执的先后顺序可能不同。示例只模拟成功投递，不会向真实手机发送短信。

## 连接与回执

SMPP 的 transceiver 会话可以在同一条 TCP 连接上发送 `submit_sm`、接收 `deliver_sm` 回执。示例也接受 receiver 和 transmitter 绑定，但 receiver 不能提交；transmitter 只能提交不需要成功回执的短信。本例不实现将同账号的回执路由到另一条 receiver 连接。

`registered_delivery=1` 请求成功或失败时返回回执，示例在默认 100～500 ms 后模拟成功结果；0 不请求 SMSC 回执，2 仅请求失败回执，因此模拟成功时不推送。其它确认标志不在本例范围内，提交会返回 `ESME_RINVREGDLVFLG`。

回执设置 `esm_class=0x04`，源、目标地址及 TON/NPI 与提交方向相反。正文 `DeliveryReceipt.ID` 是原 `submit_sm_resp.MessageID` 对应的文本字符串，同时携带 `receipted_message_id`（含结尾 NUL）和 `message_state=2`（DELIVERED）TLV。客户端通过 `deliver_sm_resp` 回显 **回执 PDU 的流水号**；该响应的 `message_id` 按协议置空，不能照搬 SMGP 的消息 ID 回显方式。

固定 worker 和有界队列限制后台回执任务，队列满时提交返回 `ESME_RTHROTTLED`。服务端和客户端都可发起 `enquire_link`；连续三次未应答后服务端关闭连接。流水号限制在 SMPP 规定的 `1..0x7fffffff`。

错误绑定响应会先写出再断开连接；`unbind_resp` 也在写完后关闭连接。SIGINT/SIGTERM 触发有截止时间的关闭，并取消仍在队列中的模拟回执。日志不输出密码、正文或完整手机号码。

协议依据见 [SMPP 3.4](../../../doc/标准协议_SMPP3.4.pdf) §2.2、§2.11、§4.1、§4.4、§4.6、§4.11 和 Appendix B。

## 配置与验证

```bash
go run ./nioserver/examples/smpp34server -help
go run ./nioserver/examples/smpp34client -help
go test ./nioserver/examples/smpp34...
go test -race -tags=integration ./nioserver/examples/smpp34server
```

服务端可配置 `-listen`、`-receipt-workers`、`-receipt-queue`、回执延迟、探活间隔、未应答次数和传输/关闭超时。队列容量按单条提交计数。客户端提供 `-server`、`-source`、`-destination`、`-message`，以及覆盖完整交互的 `-timeout`（默认 30 秒）；客户端的演示正文限制为 140 字节 ASCII。

这是单账号、单进程的接线示例：数据只在内存中，不提供持久化、回执重发/超时跟踪、跨连接回执路由、滑动窗口、长短信重组或真实计费。消息 ID 在进程重启后重新计数，不具备全局唯一性。
