# SMGP 3.0 客户端示例

客户端使用标准库 TCP 连接，以收发模式登录 SMGP 3.0 服务端，校验服务端鉴权摘要，提交一条要求回执的模拟短信，确认 `DELIVER` 回执、验证心跳并发送 `EXIT`。整个过程默认在 30 秒内完成。

配套服务端、启动命令、预期输出及 SGIP/SMGP 回执连接的区别见 [SMGP 示例说明](../smgp30server/README.md)。

```bash
export SMGP30_ACCOUNT='900001'
export SMGP30_PASSWORD='test-secret'
go run ./nioserver/examples/smgp30client -server=127.0.0.1:8890
```
