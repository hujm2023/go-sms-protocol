# go-sms-protocol

[![Go Reference](https://pkg.go.dev/badge/github.com/hujm2023/go-sms-protocol.svg)](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol)
[![Go Report Card](https://goreportcard.com/badge/github.com/hujm2023/go-sms-protocol)](https://goreportcard.com/report/github.com/hujm2023/go-sms-protocol)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Coverage](https://img.shields.io/badge/Coverage-49.7%25-yellow)

本项目是用 Go 语言实现的短信标准协议集合，支持主流的运营商短信网关协议，包括 SMPP、CMPP、SGIP、SMGP 等，适用于短信网关、SP、ISMG 等场景。项目结构清晰，易于扩展和维护，适合企业级短信平台、聚合网关、协议适配器等多种场景。

---

## ✨ 特性亮点

- 支持 SMPP3.4/5.0、CMPP2.0/3.0、SGIP1.2、SMGP3.0 等主流短信协议
- 完整实现消息打包、解包、状态报告、链路检测等核心功能
- 代码结构清晰，易于扩展和维护
- 提供高性能的网络服务端组件，适合大规模并发场景
- 兼容多种编码格式，支持长短信拆分与组装

## 📦 支持协议

- [x] SMPP3.4（国际短信标准协议 3.0）
- [x] SMPP5.0（国际短信标准协议 5.0）
- [x] CMPP2.0（中国移动 2.0）
- [x] CMPP3.0（中国移动 3.0）
- [x] SGIP1.2（中国联通）
- [x] SMGP3.0（中国电信）

## 📁 目录结构

- `cmpp/`：CMPP 协议实现（含 2.0、3.0）
- `sgip/`：SGIP 协议实现（含 1.2）
- `smgp/`：SMGP 协议实现（含 3.0）
- `smpp/`：SMPP 协议实现（含 3.4、5.0）
- `codec/`、`datacoding/`：协议通用的消息编解码与编码格式支持
- `nioserver/`：高性能网络服务端组件，可帮助你快速构建高性能服务网关
- `packet/`：二进制数据包编解码工具
- `doc/`：各协议官方标准文档（PDF）

## 🚀 安装方式

- 需 Go 1.18 及以上版本
- 依赖详见 go.mod

```shell
 go get github.com/hujm2023/go-sms-protocol
```

## 🛠️ 快速开始

### 1. 克隆仓库

```shell
git clone https://github.com/hujm2023/go-sms-protocol.git
cd go-sms-protocol
```

### 2. 基础功能

- 最优编码选择
- 长短信拼接
- 长短信拆分
- 各协议编解码支持

详细用法请参考各协议目录下的测试用例。

## 🙋 常见问题（FAQ）

- **Q: 如何支持长短信拆分与组装？**
  A: 本库已内置长短信拆分与组装逻辑，详见 `longsms.go` 及相关协议实现。
- **Q: 如何自定义编码格式？**
  A: 可通过 `datacoding/` 目录下的编码器进行扩展或自定义。

## 🤝 社区与贡献

- 欢迎提交 issue、PR 及建议
- 代码需遵循 Go 语言规范，建议补充必要的注释和测试
- 详细贡献流程见 [CONTRIBUTING.md]（如有）

## 📚 参考文档

- 项目根目录 doc/ 下包含各协议官方标准 PDF，可供详细查阅

## 🏗️ 可扩展性与架构设计

本项目高度重视可扩展性与模块解耦，便于二次开发和协议适配：

- **协议适配层**：各主流短信协议（SMPP、CMPP、SGIP、SMGP）均采用独立目录和模块实现，遵循统一接口规范，便于新增或替换协议实现。
- **插件化机制**：核心功能与协议实现解耦，支持通过接口扩展自定义消息处理、编码格式、链路管理等，满足多样化业务需求。
- **灵活的消息编解码**：`codec/` 和 `datacoding/` 目录下实现了通用的消息编解码框架，支持多种编码格式（如 GSM7、UCS2、GB18030 等），可按需扩展。
- **高性能网络服务端**：`nioserver/` 提供基于事件驱动的高性能网络服务端组件，支持大规模并发连接，适合服务网关和运营级应用。
- **模块解耦与可测试性**：各协议、工具、网络层均为独立包，便于单元测试和功能扩展，提升代码可维护性。
- **易于集成与定制**：通过接口和配置，开发者可快速集成本库至自有系统，或根据业务场景定制协议细节和消息处理逻辑。

该架构设计确保了项目的灵活性、可维护性和易用性，适合企业级短信平台、聚合网关、协议适配器等多种场景。

### 📖 扩展指南

- **添加新协议**：实现 protocol.PDU 接口以适配新协议消息类型；如需特殊的数据包长度判定，可实现对应的 codec.Codec；并新增协议专属的解码分发器（如 DecodeCMPP30）。
- **添加新 PDU**：在对应协议/版本包下定义新的 PDU 结构体，实现 protocol.PDU 接口，并在协议的解码分发函数中注册。
- **添加新编码方式**：实现 datacoding.Codec 接口，并根据需要在 codec_cmpp.go、codec_smpp.go 等协议包装器中注册。
- **自定义服务端行为**：通过 nioserver.BaseServer 的 ServerOption 选项自定义日志、工作池、连接生命周期回调（如 OnCloseFunc、WithRefreshCtxWhenRead）；可用 ISMSConn.SetBizData 为连接附加自定义业务数据。

上述机制可帮助开发者灵活扩展协议、消息类型和编码方式，并根据业务需求定制服务端行为。

## 📬 联系方式

- 作者：hujm2023
- Email: <hujm2023@gmail.com>
