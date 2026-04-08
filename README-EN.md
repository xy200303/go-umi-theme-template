# Enterprise Website Template

An enterprise-ready full-stack starter built with `Gin + Umi 4 + React + Casbin + JWT + PostgreSQL + Redis`. It includes authentication, RBAC permissions, system configuration, audit logging, and a unified file center, making it a solid base for admin dashboards and enterprise portals.

<img width="2490" height="1477" alt="image" src="https://github.com/user-attachments/assets/d9ce0569-e061-4a8c-a632-cd8f56e2ff5c" />

<img width="2500" height="1465" alt="image" src="https://github.com/user-attachments/assets/e75466f4-2a60-4c78-82bd-b7f3f69d429a" />

<img width="2494" height="1451" alt="image" src="https://github.com/user-attachments/assets/b7c7f4b8-eaf0-4618-aaee-9e218864d398" />

<img width="2467" height="1472" alt="image" src="https://github.com/user-attachments/assets/1b2e618e-9309-4601-b4f8-e756483ad06e" />

## Features

- Authentication: password login, SMS-code login, auto-login after registration, and Access Token / Refresh Token refresh flow
- Frontend permission control: dynamically renders navbar items, admin menus, and page entry points based on the current user's permissions
- Profile center: profile editing, avatar upload, password change, and phone rebinding
- System management: includes users, roles, system configs, file management, and audit logs
- Role policies: generates `openapi.json` from controller annotations, and the frontend renders policy selection from `operationId`
- Audit logging: records user API operations through middleware, with filters by keyword, module, and status code
- File center: unified upload, direct-upload initialization, direct-upload completion, and signed download; supports local storage and Tencent COS
- Internationalization: built-in `zh-CN` / `en-US`
- Frontend engineering: based on Umi 4 directory-based pages with Tailwind CSS v4, Ant Design, and Zustand

## Tech Stack

- Backend: Go, Gin, Gorm, Casbin, Redis, PostgreSQL
- Frontend: Umi 4, React 19, TypeScript, Ant Design, Tailwind CSS v4, Zustand, Axios
- Build tooling: Bun
- Cloud services: Tencent SMS, Tencent COS

## Project Structure

```text
backend/
  cmd/server                    # Go service entry point
  configs/                      # Casbin configs and other runtime files
  generate/                     # OpenAPI permission template generator and embedded assets
  internal/
    api/                        # controllers, middleware, routes
    models/
      dto/                      # request / response DTOs
      entities/                 # Gorm entities
      mapper/                   # model mappers
    pkg/                        # config, database, cache, utils
    repository/                 # data access layer grouped by business domain
    service/                    # auth, access, user, file, admin services
  web/                          # frontend build output

frontend/
  src/
    api/                        # API wrappers
    components/                 # reusable components
    constants/                  # routes and shared constants
    i18n/                       # internationalization
    lib/                        # utilities
    pages/                      # Umi directory-based pages
      admin/
        home/
        system/
          audit/
          config/
          files/
          role/
          users/
```

## Main Pages

- `/`: home page
- `/about`: about page
- `/blog`: news page
- `/login`: login page
- `/register`: registration page
- `/profile`: profile center
- `/admin/home`: admin dashboard
- `/admin/system/users`: user management
- `/admin/system/role`: role management
- `/admin/system/config`: system configuration
- `/admin/system/files`: file management
- `/admin/system/audit`: audit logs

## Quick Start

### Option 1: Local Development

1. Start PostgreSQL and Redis

```bash
docker compose up -d postgres redis
```

Default port mappings:

- PostgreSQL: `127.0.0.1:5433`
- Redis: `127.0.0.1:6380`

2. Configure backend environment variables

Copy `backend/.env.example` to `backend/.env`:

```bash
cd backend
copy .env.example .env
```

A typical local setup looks like this:

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

3. Start the backend

```bash
cd backend
go run ./cmd/server
```

Default address: `http://127.0.0.1:8080`

4. Start the frontend

```bash
cd frontend
bun install
bun run dev
```

Default address: `http://127.0.0.1:5173`

In development, Umi proxies `/api` to `http://127.0.0.1:8080`.

### Option 2: One-Command Docker Compose Startup

```bash
docker compose up -d --build
```

By default this starts:

- `app`: Go service with frontend static assets on port `8080`
- `postgres`: PostgreSQL 16 mapped to host port `5433`
- `redis`: Redis 7 mapped to host port `6380`

After startup, open:

- Application: `http://127.0.0.1:8080`

## Frontend Build

```bash
cd frontend
bun run build
```

The build output is written to `backend/web` by default. The Go backend serves the static files directly and handles SPA fallback routing.

## Environment Variables

Reference file: `backend/.env.example`

### Core Settings

- `SERVER_PORT`: backend port
- `SERVER_MODE`: runtime mode, usually `debug` or `release`
- `FRONTEND_DIST_DIR`: frontend build directory
- `POSTGRES_DSN`: PostgreSQL connection string
- `REDIS_ADDR`: Redis address
- `REDIS_PASSWORD`
- `REDIS_DB`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `JWT_ACCESS_EXPIRE_MIN`
- `JWT_REFRESH_EXPIRE_DAY`

### SMS Settings

- `SMS_VERIFY_ENABLED`: enables or disables SMS verification features
- `TENCENT_SMS_SECRET_ID`
- `TENCENT_SMS_SECRET_KEY`
- `TENCENT_SMS_SDK_APP_ID`
- `TENCENT_SMS_SIGN_NAME`
- `TENCENT_SMS_TEMPLATE_ID`
- `TENCENT_SMS_REGION`

The frontend reads `GET /api/v1/auth/options` to determine whether to show SMS login, registration verification codes, and phone-rebinding verification steps.

### Upload Settings

- `UPLOAD_DRIVER`: `auto` / `local` / `cos`
- `UPLOAD_LOCAL_PATH`: local upload directory
- `UPLOAD_MAX_SIZE_MB`
- `UPLOAD_ALLOWED_SUFFIX`
- `TENCENT_COS_SECRET_ID`
- `TENCENT_COS_SECRET_KEY`
- `TENCENT_COS_BUCKET_URL`
- `TENCENT_COS_BASE_URL`

Notes:

- In local mode, files are received and stored by the backend
- In COS mode, direct-upload initialization and completion callbacks are supported
- All downloads go through the signed download endpoint

### Initial Admin

- `INIT_ADMIN_USERNAME`
- `INIT_ADMIN_PHONE`
- `INIT_ADMIN_PASSWORD`

## System Config Conventions

Important built-in system configs currently include:

- `audit.max_records`: maximum retained audit log count, default limit `10000`

## Permission Template Generation

Role policy templates are generated automatically from controller annotations instead of being maintained by hand:

```go
// ListUsers godoc
// @Summary List users
// @Description Allows viewing and searching the user list
// @Tags users
// @ID users.list
// @Router /api/v1/admin/users [get]
func (ctl *AdminController) ListUsers(c *gin.Context) {}
```

Generate them with:

```bash
cd backend
go generate ./generate
```

Output file:

```text
backend/generate/openapi.json
```

At runtime, the backend reads the `operationId` list, and the admin role page renders grouped policy options from it.

## API Overview

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

## Common Commands

```bash
# install frontend dependencies
make deps

# start backend
make backend

# start frontend dev server
make frontend

# build frontend and backend
make build

# clean frontend build output
make clean
```

## Testing and Build

```bash
# backend tests
cd backend
go test ./...

# backend build
cd backend
go build ./...

# regenerate permission templates
cd backend
go generate ./generate

# frontend type check
cd frontend
bun run typecheck

# frontend build
cd frontend
bun run build
```

## Notes

- The current frontend is a Umi 4 + React CSR application, statically served by the Go backend after build
- Admin menus and frontend navigation are rendered dynamically based on the current user's permissions
- When SMS verification is enabled, registration, SMS login, and phone rebinding automatically switch to verification-code flows
