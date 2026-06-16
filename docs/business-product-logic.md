# sub2api 业务与产品逻辑

## 这是什么

sub2api 是订阅额度分发型 AI API Gateway。它把上游账号能力、计费、并发和路由控制封装成一个对外 API key 平台。

## 核心产品逻辑

- 上游资源来自多账号、多认证方式；下游用户只看到平台发放的 API key。
- 真正的产品价值在调度、配额、计费、粘性会话和并发控制，而不是简单代理转发。
- 管理后台和外部系统集成面向运营与商业化，API 网关面向开发者与终端消费。

## 核心业务闭环

1. 管理员接入上游账号资源。
2. 平台为用户生成 API key 与额度策略。
3. 用户请求进入网关。
4. 调度器选择合适上游账号并维持 sticky session / concurrency / rate limit。
5. 平台记录 usage、billing 和审计数据。
6. 管理后台和支付/工单等外部系统辅助运营。

## 当前业务边界

- 这是网关与运营平台，不是单一模型 provider SDK。
- 上游账号管理和下游 key 分发都属于核心业务，不应被误写成“普通反向代理”。
- README 里强调官方域名与 demo/部署方式，说明该 repo 已经有独立产品面。

## 推荐阅读顺序

1. `README.md`
2. `README_CN.md`
3. `docs/ADMIN_PAYMENT_INTEGRATION_API.md`
4. `deploy/README.md`

