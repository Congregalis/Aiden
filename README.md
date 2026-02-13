# Aiden

Aiden 是一个面向 Telegram 的对话式自学助手。当前仓库已完成 M1 的工程初始化骨架。

## 本地开发

### 1. 准备环境变量

```bash
cp configs/.env.example .env
```

请至少填写以下必填项：

- `DB_DSN`
- `TELEGRAM_BOT_TOKEN`

数据库连接池参数可按需调整：

- `DB_MAX_OPEN_CONNS`（默认 `20`）
- `DB_MAX_IDLE_CONNS`（默认 `10`）
- `DB_CONN_MAX_LIFETIME`（默认 `30m`）

### 2. 启动依赖（PostgreSQL）

```bash
docker compose up -d postgres
```

### 3. 启动应用

```bash
make run
```

启动后可检查：

- `GET http://localhost:8080/healthz`
- `GET http://localhost:8080/readyz`

## 常用命令

```bash
make test
make lint
make migrate-up
make migrate-down
make migrate-seed
```

`migrate-up` / `migrate-down` 依赖环境变量 `DB_DSN`，并按 `migrations/` 下 SQL 文件顺序执行。
`migrate-seed` 可选执行 `migrations/seed_m1.sql` 初始化本地演示数据。

## 使用 Docker Compose 一键启动

```bash
docker compose up --build
```

默认会启动：

- `app`：Aiden 服务
- `postgres`：PostgreSQL 16
