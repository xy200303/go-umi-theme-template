# 企业级网站模板

基于 `Gin + Umi 4 + React + Casbin + JWT + PostgreSQL + Redis` 的企业级全栈模板，内置认证、RBAC 权限、系统配置、日志审计和统一文件中心，适合作为后台管理系统或企业门户的二开基础。

<img width="2490" height="1477" alt="image" src="https://github.com/user-attachments/assets/d9ce0569-e061-4a8c-a632-cd8f56e2ff5c" />

<img width="2500" height="1465" alt="image" src="https://github.com/user-attachments/assets/e75466f4-2a60-4c78-82bd-b7f3f69d429a" />

<img width="2494" height="1451" alt="image" src="https://github.com/user-attachments/assets/b7c7f4b8-eaf0-4618-aaee-9e218864d398" />

<img width="2467" height="1472" alt="image" src="https://github.com/user-attachments/assets/1b2e618e-9309-4601-b4f8-e756483ad06e" />

## 功能特性

- 认证体系：支持密码登录、短信验证码登录、注册后自动登录、Access Token / Refresh Token 刷新
- 前端权限控制：根据当前用户权限动态渲染导航栏、后台菜单和页面入口
- 个人中心：支持资料维护、头像上传、密码修改、手机号换绑
- 系统管理：提供用户、角色、系统配置、文件管理、日志审计等后台能力
- 角色策略：从控制器注释生成 `openapi.json`，前端基于 `operationId` 进行策略勾选与展示
- 日志审计：通过中间件记录用户接口操作，支持按关键词、模块、状态码筛选
- 文件中心：统一文件上传、直传初始化、直传完成、签名下载；支持本地存储和腾讯云 COS
- 国际化：内置 `zh-CN` / `en-US`
- 前端工程：基于 Umi 4 目录式页面结构，使用 Tailwind CSS v4、Ant Design 和 Zustand

## 技术栈

- 后端：Go、Gin、Gorm、Casbin、Redis、PostgreSQL
- 前端：Umi 4、React 19、TypeScript、Ant Design、Tailwind CSS v4、Zustand、Axios
- 构建工具：Bun
- 云服务：Tencent SMS、Tencent COS

## 项目结构

```text
backend/
  cmd/server                    # Go 服务入口
  configs/                      # Casbin 配置等
  generate/                     # OpenAPI 权限模板生成器与嵌入资源
  internal/
    api/                        # controller、middleware、routes
    models/
      dto/                      # 请求/响应 DTO
      entities/                 # Gorm 实体
      mapper/                   # 模型映射
    pkg/                        # config、database、cache、utils
    repository/                 # 按业务拆分的数据访问层
    service/                    # 认证、权限、用户、文件、后台业务
  web/                          # 前端构建产物

frontend/
  src/
    api/                        # 接口封装
    components/                 # 通用组件
    constants/                  # 路由与常量
    i18n/                       # 国际化
    lib/                        # 工具函数
    pages/                      # Umi 目录式页面
      admin/
        home/
        system/
          audit/
          config/
          files/
          role/
          users/
```

## 主要页面

- `/`：首页
- `/about`：关于页
- `/blog`：资讯页
- `/login`：登录页
- `/register`：注册页
- `/profile`：个人中心
- `/admin/home`：后台首页
- `/admin/system/users`：用户管理
- `/admin/system/role`：角色管理
- `/admin/system/config`：系统配置
- `/admin/system/files`：文件管理
- `/admin/system/audit`：日志审计

## 快速开始

### 方式一：本地开发

1. 启动 PostgreSQL 和 Redis

```bash
docker compose up -d postgres redis
```

默认端口映射：

- PostgreSQL：`127.0.0.1:5433`
- Redis：`127.0.0.1:6380`

2. 配置后端环境变量

复制 `backend/.env.example` 为 `backend/.env`：

```bash
cd backend
copy .env.example .env
```

本地开发常见配置示例：

```env
SERVER_PORT=8080
SERVER_MODE=debug
FRONTEND_DIST_DIR=web
SMS_VERIFY_ENABLED=false

POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5433/enterprise_web?sslmode=disable&TimeZone=Asia/Shanghai
REDIS_ADDR=127.0.0.1:6380

JWT_ACCESS_SECRET=replace-with-access-secret
JWT_REFRESH_SECRET=replace-with-refresh-secret
```

3. 启动后端

```bash
cd backend
go run ./cmd/server
```

默认地址：`http://127.0.0.1:8080`

4. 启动前端

```bash
cd frontend
bun install
bun run dev
```

默认地址：`http://127.0.0.1:5173`

开发环境中，Umi 已将 `/api` 代理到 `http://127.0.0.1:8080`。

### 方式二：Docker Compose 一键启动

```bash
docker compose up -d --build
```

默认会启动：

- `app`：Go 服务 + 前端静态资源，端口 `8080`
- `postgres`：PostgreSQL 16，宿主机端口 `5433`
- `redis`：Redis 7，宿主机端口 `6380`

启动完成后访问：

- 应用地址：`http://127.0.0.1:8080`

## 前端构建

```bash
cd frontend
bun run build
```

构建产物默认输出到 `backend/web`，Go 后端会直接托管静态资源并处理 SPA 回退。

## 环境变量

参考文件：`backend/.env.example`

### 核心配置

- `SERVER_PORT`：后端端口
- `SERVER_MODE`：运行模式，常见为 `debug` 或 `release`
- `FRONTEND_DIST_DIR`：前端构建目录
- `POSTGRES_DSN`：PostgreSQL 连接串
- `REDIS_ADDR`：Redis 地址
- `REDIS_PASSWORD`
- `REDIS_DB`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `JWT_ACCESS_EXPIRE_MIN`
- `JWT_REFRESH_EXPIRE_DAY`

### 短信配置

- `SMS_VERIFY_ENABLED`：是否启用短信验证码能力
- `TENCENT_SMS_SECRET_ID`
- `TENCENT_SMS_SECRET_KEY`
- `TENCENT_SMS_SDK_APP_ID`
- `TENCENT_SMS_SIGN_NAME`
- `TENCENT_SMS_TEMPLATE_ID`
- `TENCENT_SMS_REGION`

当前前端会通过 `GET /api/v1/auth/options` 获取短信开关状态，并据此动态显示验证码登录、注册验证码和换绑手机号验证码步骤。

### 上传配置

- `UPLOAD_DRIVER`：`auto` / `local` / `cos`
- `UPLOAD_LOCAL_PATH`：本地上传目录
- `UPLOAD_MAX_SIZE_MB`
- `UPLOAD_ALLOWED_SUFFIX`
- `TENCENT_COS_SECRET_ID`
- `TENCENT_COS_SECRET_KEY`
- `TENCENT_COS_BUCKET_URL`
- `TENCENT_COS_BASE_URL`

说明：

- 本地模式下文件由服务端接收并存储
- COS 模式下支持直传初始化和完成回写
- 文件下载统一走签名下载接口

### 初始管理员

- `INIT_ADMIN_USERNAME`
- `INIT_ADMIN_PHONE`
- `INIT_ADMIN_PASSWORD`

## 系统配置约定

当前内置的重要系统配置包括：

- `audit.max_records`：日志审计最大保留条数，默认限制为 `10000`

## 权限模板生成

角色策略模板不是手工维护，而是由控制器注释自动生成：

```go
// ListUsers godoc
// @Summary 查看用户列表
// @Description 允许查看并按条件搜索用户列表
// @Tags users
// @ID users.list
// @Router /api/v1/admin/users [get]
func (ctl *AdminController) ListUsers(c *gin.Context) {}
```

生成命令：

```bash
cd backend
go generate ./generate
```

生成结果：

```text
backend/generate/openapi.json
```

运行时会由后端读取 `operationId` 列表，并在后台角色管理中按分组展示策略项。

## API 概览

### Auth

- `GET /api/v1/auth/options`
- `POST /api/v1/auth/sms/send`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login/password`
- `POST /api/v1/auth/login/sms`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`

### User

- `GET /api/v1/user/profile`
- `PUT /api/v1/user/profile`
- `POST /api/v1/user/password/reset`
- `POST /api/v1/user/phone/change`
- `POST /api/v1/user/avatar/upload`
- `POST /api/v1/user/files/upload`
- `POST /api/v1/user/files/direct/init`
- `POST /api/v1/user/files/direct/complete`

### File Download

- `GET /api/v1/files/:id/download`

### Admin

- `GET /api/v1/admin/stats`
- `GET /api/v1/admin/files`
- `GET /api/v1/admin/files/stats`
- `GET /api/v1/admin/audit-logs`
- `GET /api/v1/admin/policy-templates`
- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users`
- `PUT /api/v1/admin/users/:id`
- `DELETE /api/v1/admin/users/:id`
- `PUT /api/v1/admin/users/:id/password`
- `PUT /api/v1/admin/users/:id/roles`
- `GET /api/v1/admin/roles`
- `POST /api/v1/admin/roles`
- `PUT /api/v1/admin/roles/:id`
- `DELETE /api/v1/admin/roles/:id`
- `GET /api/v1/admin/roles/:id/policies`
- `PUT /api/v1/admin/roles/:id/policies`
- `GET /api/v1/admin/system-configs`
- `PUT /api/v1/admin/system-configs`

## 常用命令

```bash
# 安装前端依赖
make deps

# 启动后端
make backend

# 启动前端开发环境
make frontend

# 构建前后端
make build

# 清理前端构建产物
make clean
```

## 测试与构建

```bash
# 后端测试
cd backend
go test ./...

# 后端编译
cd backend
go build ./...

# 重新生成权限模板
cd backend
go generate ./generate

# 前端类型检查
cd frontend
bun run typecheck

# 前端构建
cd frontend
bun run build
```

## 说明

- 当前前端为 Umi 4 + React 的 CSR 应用，构建后由 Go 后端静态托管
- 后台菜单和前端导航会根据当前用户权限动态渲染
- 如果启用短信验证，注册、短信登录和手机号换绑会自动切换为验证码流程
