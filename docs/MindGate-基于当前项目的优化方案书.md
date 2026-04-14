# MindGate 基于当前项目的优化方案书

## 1. 方案目标

当前项目已经具备一个较完整的企业级全栈后台基础，包括用户注册登录、JWT 鉴权、刷新令牌、RBAC 权限、审计日志、系统配置、文件中心、个人中心、后台管理、PostgreSQL、Redis 和 Docker Compose 部署。

MindGate 不建议从零重写，而应该在现有项目上进行业务化改造：

- 保留现有认证、用户、角色、权限、审计、文件、系统配置能力。
- 新增 MindGate 专属的大模型网关、模型供应商、对话记录、知识库、知识图谱和请求增强模块。
- 将当前“企业级网站模板”逐步调整为“多协议 AI API 网关 + 个人知识库平台”。
- 先跑通可用闭环，再逐步强化多协议适配、隐私治理、知识沉淀和可视化能力。

## 2. 当前项目能力盘点

### 2.1 已实现能力

当前项目已经实现的基础能力如下：

- 后端服务：Go + Gin 服务入口和路由注册。
- 数据存储：PostgreSQL + Gorm AutoMigrate。
- 缓存组件：Redis，用于刷新令牌、短信验证码等场景。
- 认证模块：注册、密码登录、短信登录、刷新 Token、退出登录。
- 用户中心：个人资料、头像上传、密码修改、手机号换绑。
- 权限模块：角色、用户角色、Casbin RBAC 权限校验。
- 审计日志：中间件记录接口访问、操作名称、状态码、耗时、IP、User-Agent。
- 管理后台：用户管理、角色管理、系统配置、文件管理、审计日志。
- 文件中心：本地上传、直传初始化、直传完成、签名下载，支持本地和 COS 思路。
- 前端工程：Umi 4 + React + Ant Design + Zustand，已有登录、注册、个人中心、后台布局。
- 部署基础：Dockerfile + docker-compose，包含 app、postgres、redis。

### 2.2 当前项目可复用点

MindGate 可以直接复用：

- 用户身份体系：作为 MindGate 用户、API Token 所属主体和数据隔离依据。
- RBAC 权限体系：用于控制管理员功能、用户功能、团队空间功能。
- 审计日志体系：用于记录后台管理操作和敏感配置操作。
- 系统配置体系：用于保存全局网关策略、脱敏策略、任务参数、默认 Provider 模板。
- 文件中心：用于导出知识库、保存附件、归档长对话、上传导入文件。
- PostgreSQL：用于保存 Provider 配置、对话记录、知识卡片、任务状态。
- Redis：用于限流、异步任务队列、流式请求状态、短期缓存。
- 前端后台框架：用于快速新增 Provider、Conversation、Knowledge、Graph 等页面。

### 2.3 当前项目需要调整的问题

当前项目还是企业后台模板，MindGate 化时建议优化：

- 项目命名仍偏 `enterprise_app`、`enterprise_web`，需要逐步改为 MindGate。
- 当前路由主要集中在 `/api/v1/auth`、`/api/v1/user`、`/api/v1/admin`，还没有独立网关入口。
- 当前审计日志适合管理操作，不适合直接保存 AI 对话原文，需要新增 Conversation 数据域。
- 当前 RBAC 默认只放行 `/api/v1/user/*` 和 `/api/v1/admin/*`，需要新增 MindGate 业务权限。
- 当前系统配置是纯文本 `config_val`，敏感配置和 Provider Key 不能直接复用该表明文存储。
- 当前 Docker Compose 没有 pgvector、worker、对象存储、队列监控等 MindGate 后续组件。
- 当前前端菜单仍是系统管理导向，需要新增用户侧 MindGate 工作台。

## 3. 总体优化方向

### 3.1 架构演进原则

建议采用“保留底座 + 新增业务域 + 渐进替换品牌”的方式演进。

第一步不动认证和后台基础，新增 MindGate 模块：

```text
backend/internal/
  api/
    controllers/
    middleware/
    routes/
  models/
  repository/
  service/
```

继续沿用当前分层风格，新增：

```text
backend/internal/
  api/controllers/gateway_controller.go
  api/controllers/provider_controller.go
  api/controllers/conversation_controller.go
  api/controllers/knowledge_controller.go
  api/controllers/graph_controller.go
  api/middleware/gateway_auth_middleware.go
  service/gateway/
  service/provider/
  service/conversation/
  service/knowledge/
  service/graph/
  repository/provider/
  repository/conversation/
  repository/knowledge/
  repository/graph/
  models/entities/provider.go
  models/entities/conversation.go
  models/entities/knowledge.go
  models/entities/gateway_token.go
```

前端新增：

```text
frontend/src/pages/gateway/
frontend/src/pages/providers/
frontend/src/pages/conversations/
frontend/src/pages/knowledge/
frontend/src/pages/graph/
frontend/src/api/endpoints/gateway.ts
frontend/src/api/endpoints/provider.ts
frontend/src/api/endpoints/conversation.ts
frontend/src/api/endpoints/knowledge.ts
```

### 3.2 产品信息架构优化

建议将前端从“首页 + 关于 + 博客 + 管理后台”调整为 MindGate 产品结构：

- 首页：MindGate 产品介绍、接入方式、核心价值。
- 工作台：请求量、Token 用量、最近会话、知识卡片、异常请求。
- 我的接口：用户管理自己的 AI 转发接口。
- 接口配置：创建不同平台的转发接口，选择接口类型、填写目标 API 和目标 API Key，并生成 MindGate 网关 Key。
- 网关接入：展示 MindGate 提供的四种 API 格式、Base URL、网关 Key 和客户端配置示例。
- 对话记录：查看通过网关产生的 AI 会话。
- 知识库：查看、搜索、编辑、导出知识卡片。
- 知识图谱：展示知识之间的关系。
- 隐私设置：配置记录范围、脱敏、删除、导出。
- 系统管理：保留现有用户、角色、文件、配置、审计页面。

## 4. 后端优化方案

### 4.1 路由分层优化

当前已有 `/api/v1` 管理接口。MindGate 建议拆成三类路由：

```text
/api/v1/*                  # 管理后台和用户中心接口，继续走 JWT + RBAC
/v1/*                      # OpenAI-compatible 网关入口
/anthropic/v1/*            # Anthropic/Claude-compatible 网关入口
/gemini/v1beta/*           # Gemini-compatible 网关入口
/mg/v1/*                   # MindGate 自有统一协议入口
```

这样可以避免 AI 客户端因为路径不兼容而接入困难。

管理接口建议新增：

```text
GET    /api/v1/my-interfaces
POST   /api/v1/my-interfaces
GET    /api/v1/my-interfaces/:id
PUT    /api/v1/my-interfaces/:id
DELETE /api/v1/my-interfaces/:id
POST   /api/v1/my-interfaces/:id/test
POST   /api/v1/my-interfaces/:id/rotate-key

GET    /api/v1/conversations
GET    /api/v1/conversations/:id
DELETE /api/v1/conversations/:id

GET    /api/v1/knowledge/cards
POST   /api/v1/knowledge/search
PUT    /api/v1/knowledge/cards/:id
DELETE /api/v1/knowledge/cards/:id

GET    /api/v1/knowledge/graph
GET    /api/v1/privacy/settings
PUT    /api/v1/privacy/settings
```

网关入口建议新增：

```text
POST /v1/chat/completions
POST /v1/responses
GET  /v1/models

POST /anthropic/v1/messages
GET  /anthropic/v1/models

POST /gemini/v1beta/models/:model:generateContent
POST /gemini/v1beta/models/:model:streamGenerateContent
GET  /gemini/v1beta/models

POST /mg/v1/chat
POST /mg/v1/chat/stream
GET  /mg/v1/models
```

### 4.2 鉴权模型优化

当前后台接口使用用户登录后的 JWT，非常适合前端管理页面。AI 客户端则应该使用“接口级网关 Key”。这个 Key 不只是鉴权凭证，还用于确定当前请求应该走用户配置的哪一个转发接口。

建议新增两套鉴权：

- Web JWT：继续用于登录后台、个人中心、知识库页面。
- 接口级网关 Key：用于 Codex、Claude Code、Cursor、Chatbox 等外部客户端请求 MindGate 网关，同时绑定一个具体的用户转发接口。

接口级网关 Key 设计：

```text
mg_sk_xxxxxxxxxxxxxxxxxxxxx
```

数据库只保存哈希，不保存明文。一个网关 Key 对应用户创建的一个“我的接口”配置：

```text
gateway_keys
  id
  user_id
  interface_id
  token_hash
  token_prefix
  last_used_at
  expires_at
  enabled
  created_at
  updated_at
```

网关请求流程：

```text
Authorization: Bearer mg_sk_xxx
  -> GatewayAuthMiddleware
  -> 校验 token_hash
  -> 注入 user_id
  -> 读取 interface_id
  -> 读取用户接口配置
  -> 根据接口配置选择目标 API、目标 API Key、目标协议
  -> 转发上游模型
```

这样用户可以创建多个接口，例如：

```text
接口 A：OpenAI Api -> https://api.openai.com/v1 -> sk-openai-xxx -> mg_sk_a
接口 B：OpenAiResponse -> https://api.openai.com/v1 -> sk-openai-yyy -> mg_sk_b
接口 C：Claude -> https://api.anthropic.com -> sk-ant-xxx -> mg_sk_c
接口 D：Gemini -> https://generativelanguage.googleapis.com -> google-key-xxx -> mg_sk_d
```

客户端使用哪个 `mg_sk_`，MindGate 就转发到哪个接口配置。

### 4.3 我的接口配置优化

当前系统配置表不适合保存用户级目标 API Key。建议新增独立的“我的接口”表，并加密存储目标 API Key。

用户在个人中心进入：

```text
个人中心
  -> 我的接口
     -> 接口配置
```

每条接口配置表示一个可被外部 AI 客户端调用的转发接口。

用户创建接口时填写：

- 接口名称：例如“我的 OpenAI 主账号”“Claude 工作号”“Gemini Flash”。
- 接口类型：四选一，`Openai Api`、`OpenAiResponse`、`Claude`、`Gemini`。
- 目标 API 地址：例如 `https://api.openai.com/v1`。
- 目标 API Key：用户自己的上游模型 Key。
- 默认模型：例如 `gpt-4.1`、`claude-sonnet-4-5`、`gemini-2.5-pro`。
- 是否启用流式转发。
- 是否记录对话原文。
- 是否启用知识库增强。
- 备注。

保存后系统生成：

- MindGate 网关 Key：例如 `mg_sk_xxx`。
- OpenAI 格式 Base URL：例如 `https://your-domain.com/v1`。
- OpenAI Responses 格式 Base URL：例如 `https://your-domain.com/v1`。
- Claude 格式 Base URL：例如 `https://your-domain.com/anthropic`。
- Gemini 格式 Base URL：例如 `https://your-domain.com/gemini`。

用户可以把任意一个 Base URL 和这个 `mg_sk_xxx` 配到客户端中。网关 Key 决定最终转发到哪条接口配置；客户端使用的路径决定 MindGate 用哪种入站协议解析请求。

建议实体：

```text
user_interfaces
  id
  user_id
  name
  interface_type
  target_base_url
  target_api_key_encrypted
  default_model
  model_mapping
  capabilities
  record_metadata
  record_message_content
  enable_knowledge_extract
  enable_request_augment
  timeout_ms
  enabled
  created_at
  updated_at
```

其中 `interface_type` 支持：

```text
openai_chat        # Openai Api，对应 /v1/chat/completions
openai_responses   # OpenAiResponse，对应 /v1/responses
claude_messages    # Claude，对应 /anthropic/v1/messages
gemini_generate    # Gemini，对应 /gemini/v1beta/models/:model:generateContent
```

关键要求：

- `target_api_key_encrypted` 必须加密存储。
- 前端只展示 `sk-***abcd` 这类掩码。
- 删除接口配置时不删除历史会话，只禁用后续转发。
- 支持 test connection。
- 支持轮换 MindGate 网关 Key。
- 支持默认模型和模型名映射。

配置来源建议：

- 用户接口配置：个人使用，优先级最高。
- 系统接口模板：管理员配置目标平台模板，用户可一键引用后填写自己的 Key。
- 环境变量默认接口：部署者预置，用于本地或团队场景。

### 4.4 多协议网关优化

建议不要把网关逻辑写成单个大 Controller，而是拆成协议入口层、统一请求层、Provider 出口层。

结构：

```text
Client Request
  -> Client Protocol Adapter
  -> UnifiedChatRequest
  -> Gateway Router
  -> Optional Knowledge Augmenter
  -> Provider Adapter
  -> Upstream LLM API
  -> UnifiedChatResponse / UnifiedStreamEvent
  -> Client Protocol Adapter
  -> Client Response
```

建议接口：

```text
type ClientAdapter interface {
  ParseChatRequest(ctx, request) (UnifiedChatRequest, error)
  WriteChatResponse(ctx, UnifiedChatResponse) error
  WriteStreamEvent(ctx, UnifiedStreamEvent) error
}

type ProviderAdapter interface {
  Chat(ctx, UnifiedChatRequest) (UnifiedChatResponse, error)
  StreamChat(ctx, UnifiedChatRequest) (<-chan UnifiedStreamEvent, error)
  ListModels(ctx) ([]ModelInfo, error)
}
```

首批适配：

- OpenAI Client Adapter。
- Anthropic Client Adapter。
- Gemini Client Adapter。
- OpenAI Provider Adapter。
- Anthropic Provider Adapter。
- Gemini Provider Adapter。
- Ollama Provider Adapter。

### 4.5 对话记录优化

当前审计日志不应承担对话存储职责。建议新增对话域，审计日志只记录管理和安全操作。

建议实体：

```text
conversations
  id
  user_id
  title
  source_client
  source_protocol
  project_name
  summary
  tags
  privacy_level
  created_at
  updated_at

conversation_messages
  id
  conversation_id
  user_id
  role
  content
  content_redacted
  content_type
  tool_call_id
  metadata
  created_at

gateway_requests
  id
  user_id
  provider_id
  conversation_id
  gateway_token_id
  request_path
  source_protocol
  target_protocol
  model
  stream
  status_code
  prompt_tokens
  completion_tokens
  total_tokens
  latency_ms
  error_code
  error_message
  created_at
```

记录策略：

- 默认记录元数据。
- 原文记录需要用户主动开启。
- 可按 Gateway Token、Provider、客户端、项目设置记录策略。
- 流式响应边转发边缓冲，保存失败不影响主链路。

### 4.6 隐私和安全优化

MindGate 的核心数据比普通后台更敏感，因此建议新增隐私设置表：

```text
privacy_settings
  user_id
  record_metadata
  record_message_content
  enable_redaction
  enable_knowledge_extract
  enable_request_augment
  retention_days
  created_at
  updated_at
```

必须实现：

- Provider API Key 加密存储。
- Gateway Token 只存哈希。
- Authorization、Cookie、API Key 禁止进入日志。
- 对话内容支持删除、导出、停止采集。
- 自动脱敏规则覆盖常见 API Key、Bearer Token、JWT、邮箱、手机号。
- 网关入口不要接入现有 AuditLogMiddleware 记录请求体，避免泄露对话原文。

建议新增环境变量：

```text
MINDGATE_KEY_ENCRYPTION_SECRET=replace-with-32-byte-secret
MINDGATE_DEFAULT_RECORD_METADATA=true
MINDGATE_DEFAULT_RECORD_CONTENT=false
MINDGATE_MAX_STREAM_CAPTURE_BYTES=1048576
MINDGATE_REQUEST_TIMEOUT_SEC=120
```

### 4.7 知识库与异步任务优化

当前 Redis 已可复用。建议第一阶段用数据库任务表或 Redis Stream，避免引入太多新组件。

建议任务：

```text
knowledge_tasks
  id
  user_id
  task_type
  source_id
  status
  attempts
  error_message
  run_after
  created_at
  updated_at
```

任务类型：

- 会话摘要。
- 知识点抽取。
- 标签分类。
- 向量生成。
- 知识卡片生成。
- 导出文件生成。

知识卡片：

```text
knowledge_cards
  id
  user_id
  title
  summary
  content
  category
  tags
  source_conversation_id
  importance_score
  confidence_score
  embedding
  created_at
  updated_at
```

如果短期不引入 pgvector，可以先做 PostgreSQL 全文搜索和标签检索；当知识卡片稳定后再引入 pgvector。

## 5. 前端优化方案

### 5.1 路由与菜单优化

当前前端已经有路由常量和后台 Layout。建议新增用户侧 MindGate 路由：

```text
/dashboard
/profile/interfaces
/profile/interfaces/:id
/gateway/setup
/conversations
/conversations/:id
/knowledge
/knowledge/:id
/graph
/privacy
```

后台管理保留：

```text
/admin/home
/admin/system/users
/admin/system/role
/admin/system/config
/admin/system/files
/admin/system/audit
```

新增后台 MindGate 管理：

```text
/admin/mindgate/providers
/admin/mindgate/requests
/admin/mindgate/tasks
/admin/mindgate/settings
```

### 5.2 页面优先级

第一批页面建议：

- 我的接口页：展示用户创建的所有转发接口。
- 接口配置页：新增、编辑、测试、删除接口配置，生成或轮换网关 Key。
- 网关接入页：展示四种 API 格式 Base URL、网关 Key、常见客户端配置。
- 对话记录页：列表、筛选、详情、删除、生成摘要。
- 知识库页：卡片列表、搜索、分类、导出。
- 隐私设置页：记录开关、脱敏开关、数据导出、数据清理。

第二批页面建议：

- 工作台统计。
- 知识图谱。
- 请求增强配置。
- 模型路由策略。
- 团队空间。

### 5.3 UI 文案优化

当前 README 和 Docker Compose 仍带企业模板痕迹。建议统一产品命名：

- `enterprise_app` 改为 `mindgate_app`。
- `enterprise_postgres` 改为 `mindgate_postgres`。
- `enterprise_redis` 改为 `mindgate_redis`。
- 数据库名 `enterprise_web` 改为 `mindgate`。
- 首页、导航、README 改为 MindGate 产品叙事。

## 6. 数据库迁移优化方案

当前项目使用 Gorm AutoMigrate。MindGate 初期可以继续使用 AutoMigrate，但随着数据重要性提升，建议引入明确迁移机制。

短期：

- 在 `AutoMigrateAndSeed` 中加入 MindGate 新实体。
- 在默认角色策略中加入 MindGate 用户接口权限。
- 在系统配置 seed 中加入 MindGate 默认配置。

中期：

- 引入 migration 文件，避免生产环境 AutoMigrate 不可控。
- 对 Provider Key、Gateway Token、Conversation 等敏感表建立索引和约束。
- 为大表设计归档策略。

建议索引：

```text
user_interfaces(user_id, enabled)
gateway_keys(user_id, enabled)
gateway_requests(user_id, created_at)
gateway_requests(conversation_id)
conversations(user_id, updated_at)
conversation_messages(conversation_id, created_at)
knowledge_cards(user_id, category, updated_at)
knowledge_tasks(status, run_after)
```

## 7. Docker 与部署优化方案

当前 Docker Compose 已有 app、postgres、redis。MindGate 建议分阶段升级。

### 7.1 第一阶段

继续使用当前三容器：

```text
app
postgres
redis
```

调整：

- 容器名改为 MindGate。
- 数据库名改为 MindGate。
- 增加 Key 加密密钥环境变量。
- 增加网关超时、记录策略、脱敏策略配置。

### 7.2 第二阶段

增加 worker：

```text
app       # HTTP 服务和网关转发
worker    # 知识总结、向量化、导出等异步任务
postgres
redis
```

这样可以避免知识处理阻塞网关请求。

### 7.3 第三阶段

根据需要增加：

- pgvector：可以直接作为 PostgreSQL 扩展。
- MinIO：本地对象存储。
- Qdrant：独立向量数据库。
- Prometheus/Grafana：监控请求量、延迟、错误率。

## 8. 权限策略优化

当前默认用户角色只允许 `/api/v1/user/*`。MindGate 需要新增用户权限：

```text
/api/v1/my-interfaces/*
/api/v1/conversations/*
/api/v1/knowledge/*
/api/v1/privacy/*
```

管理员权限新增：

```text
/api/v1/admin/mindgate/*
```

注意：

- 网关入口 `/v1/*`、`/anthropic/*`、`/gemini/*`、`/mg/*` 不建议走 Casbin RBAC。
- 网关入口应该走 Gateway Token 鉴权和用户级策略。
- 后台管理接口继续走 JWT + RBAC。

## 9. 审计与日志优化

当前 AuditLogMiddleware 适合记录后台操作，但 MindGate 网关请求量大、内容敏感，不建议直接复用审计表保存网关详情。

建议拆分：

- `audit_logs`：后台管理操作、安全操作、配置变更。
- `gateway_requests`：AI 请求元数据、耗时、Token、错误。
- `conversation_messages`：用户授权后的对话内容。

审计日志增加的操作：

- 创建、修改、删除接口配置。
- 生成、轮换、删除网关 Key。
- 开启或关闭原文记录。
- 导出或删除个人数据。
- 管理员修改全局 MindGate 配置。

## 10. 分阶段实施计划

### Phase 1：品牌和基础配置改造

目标：把企业模板明确改造成 MindGate 项目。

任务：

- 更新 README、首页、导航和 Docker Compose 命名。
- 增加 MindGate 配置项和环境变量。
- 增加 MindGate 用户菜单和路由骨架。
- 保留现有登录注册、个人中心、后台系统。

交付：

- 用户登录后能看到 MindGate 工作台入口。
- 管理员仍可使用原有系统管理能力。

### Phase 2：我的接口与网关 Key

目标：用户可以配置上游模型和生成客户端调用凭证。

任务：

- 新增我的接口实体、Repository、Service、Controller。
- 新增网关 Key 实体和鉴权中间件。
- 新增我的接口页和接口配置页。
- 新增 Gateway 接入指南页。
- API Key 加密存储，Token 哈希存储。

交付：

- 用户能创建不同平台的转发接口。
- 每条接口可选择 `Openai Api`、`OpenAiResponse`、`Claude`、`Gemini` 四种类型。
- 用户能为每条接口生成 `mg_sk_` 网关 Key。
- 前端能展示四种 API 格式的客户端配置示例。

### Phase 3：多协议网关转发

目标：真实 AI 客户端可以通过 MindGate 请求模型。

任务：

- 实现 OpenAI-compatible `/v1/chat/completions`。
- 实现 Anthropic-compatible `/anthropic/v1/messages`。
- 实现 Gemini-compatible `generateContent`。
- 实现流式转发。
- 实现模型列表接口。
- 实现错误格式归一化。
- 实现请求元数据记录。

交付：

- Codex、Cursor、Claude Code、Gemini CLI 至少跑通部分客户端。
- 网关请求不依赖用户登录态，而依赖 Gateway Token。

### Phase 4：对话记录

目标：用户可以在后台看到通过网关产生的 AI 会话。

任务：

- 新增 Conversation、Message、GatewayRequest 表。
- 实现会话聚合和消息保存。
- 实现对话列表、详情、筛选、删除。
- 实现隐私设置：是否记录原文、是否脱敏。

交付：

- 用户可以按时间、模型、客户端查看 AI 使用记录。
- 用户可以关闭原文记录。

### Phase 5：知识库沉淀

目标：把对话变成可复用知识。

任务：

- 新增 KnowledgeCard、KnowledgeTask。
- 实现会话摘要。
- 实现知识点抽取。
- 实现分类、标签、重要性评分。
- 实现知识库列表、搜索、编辑、导出。

交付：

- 用户可以从聊天记录自动生成知识卡片。
- 用户可以搜索和导出个人知识库。

### Phase 6：请求增强与知识图谱

目标：让历史知识反哺未来请求。

任务：

- 实现知识检索。
- 实现请求增强开关。
- 实现上下文注入。
- 实现图谱节点和关系抽取。
- 实现图谱可视化页面。

交付：

- MindGate 可以自动检索相关知识并增强请求。
- 用户可以看到知识之间的关联。

## 11. 当前项目改造优先级

建议优先级：

1. 先改品牌和导航，让项目从视觉和文档上变成 MindGate。
2. 再做“我的接口”和网关 Key，这是网关成立的前置条件。
3. 然后做 OpenAI/Anthropic/Gemini 三个入口的最小文本转发。
4. 再做请求记录和对话页面。
5. 最后做知识库、图谱和请求增强。

不建议一开始就做：

- 完整知识图谱。
- 团队空间。
- 复杂计费。
- 所有模型协议的全量高级能力。
- 高复杂度向量数据库集成。

## 12. 关键验收标准

第一阶段验收：

- 登录注册、个人中心、后台管理保持可用。
- 项目文案、容器名、数据库名、首页叙事调整为 MindGate。
- 用户登录后可以进入 MindGate 工作台。

第二阶段验收：

- 用户可以在个人中心的“我的接口 -> 接口配置”中新增接口。
- 每条接口可以选择 `Openai Api`、`OpenAiResponse`、`Claude`、`Gemini` 四种类型之一。
- 用户可以填写目标 API 和目标 API Key。
- 目标 API Key 加密存储。
- 用户可以创建或轮换该接口对应的网关 Key。
- 网关 Key 明文只在创建时展示一次。

第三阶段验收：

- OpenAI-compatible 客户端可以通过 MindGate 转发请求。
- Anthropic-compatible 客户端可以通过 MindGate 转发请求。
- Gemini-compatible 客户端可以通过 MindGate 转发请求。
- 普通响应和流式响应都可用。
- 请求失败时返回客户端可理解的错误格式。

第四阶段验收：

- 用户可以看到网关请求记录。
- 用户授权后可以看到会话消息。
- 用户可以删除会话和关闭原文记录。

第五阶段验收：

- 系统可以从会话生成摘要。
- 系统可以生成知识卡片。
- 用户可以搜索、编辑、删除、导出知识卡片。

## 13. 总结

当前项目已经具备 MindGate 所需的用户、权限、后台、审计、配置、文件、数据库和部署底座。最优路径不是推翻重做，而是把它作为 MindGate 的管理平台和用户中心，在此基础上新增多协议 AI 网关和个人知识库业务域。

推荐近期目标：

1. 保留现有登录注册和后台能力。
2. 新增“我的接口”配置和网关 Key。
3. 实现多协议网关最小闭环。
4. 再沉淀对话记录和知识库。
