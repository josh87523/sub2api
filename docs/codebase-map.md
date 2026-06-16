# sub2api Codebase Map

> 用于人和 AI 快速定位产品逻辑、运行逻辑和对应代码入口。先读 `docs/business-product-logic.md`，再按下表跳到具体模块。

## Module Entrypoints

| Area | Entry | Notes |
|---|---|---|
| 业务逻辑 | `docs/business-product-logic.md` | AI API gateway 产品边界 |
| 项目入口 | `README.md, README_CN.md, README_JA.md` | 产品概述、部署和多语言说明 |
| 后端 | `backend/` | 账号、网关、计费、调度和 API 实现 |
| 前端 | `frontend/` | 管理后台和用户界面 |
| 部署 | `deploy/, Dockerfile, Makefile` | Docker、部署和发布入口 |
| 文档 | `docs/` | 支付、管理 API 和部署说明 |
| 工具 | `tools/` | 维护和辅助脚本 |

## Reading Contract

- 先用 `docs/business-product-logic.md` 判断这个仓库解决什么问题。
- 再按本页定位到代码或运行入口，避免从文件名猜产品逻辑。
- dated plan、archive、runtime data、generated output 只能当证据或历史参考，不能直接当当前运行合同。
- 涉及外部平台、账号、发布、支付、浏览器 profile 或凭证时，必须读真实运行面和权限边界，不能只读源码。

## Recommended First Read

1. `docs/business-product-logic.md`
2. `README.md`
3. `README_CN.md`
4. `backend/`
5. `frontend/`
