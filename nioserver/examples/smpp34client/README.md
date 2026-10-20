# SMPP 3.4 客户端示例

使用 `bind_transceiver` 登录 SMSC，提交一条请求回执的模拟短信，确认 `deliver_sm`、验证 `enquire_link`，最后完成 `unbind`。默认连接 `127.0.0.1:2775`，整个演示默认在 30 秒内完成。

```bash
export SMPP34_ACCOUNT='demo'
export SMPP34_PASSWORD='secret34'
go run ./nioserver/examples/smpp34client
```

先启动配套服务端；完整命令、字段限制和回执说明见 [SMPP 示例说明](../smpp34server/README.md)。
