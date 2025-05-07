<!-- @format -->
![Coverage](https://img.shields.io/badge/Coverage-49.7%25-yellow)

# go-sms-protocol

[![Go Reference](https://pkg.go.dev/badge/github.com/hujm2023/go-sms-protocol.svg)](https://pkg.go.dev/github.com/hujm2023/go-sms-protocol)
[![Go Report Card](https://goreportcard.com/badge/github.com/hujm2023/go-sms-protocol)](https://goreportcard.com/report/github.com/hujm2023/go-sms-protocol)
[![Coverage Status](https://coveralls.io/repos/github/hujm2023/go-sms-protocol/badge.svg?branch=main)](https://coveralls.io/github/hujm2023/go-sms-protocol?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

本项目是用 Go 语言实现的短信标准协议集合，支持主流的运营商短信网关协议，包括 SMPP、CMPP、SGIP、SMGP 等，适用于短信网关、SP、ISMG 等场景。

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
# 推荐使用 go get 方式集成
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

- 2.2 长短信拼接

- 2.3 长短信拆分

- 2.4 各协议编解码支持

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

## 📬 联系方式

- 作者：hujm2023
- Email: <hujm2023@gmail.com>
