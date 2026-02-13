# Aiden M1 研发 Todo List（细粒度）

## 0. 跟踪约定

- 状态使用：`TODO` / `DOING` / `BLOCKED` / `DONE`
- 勾选规则：完成后将 `[ ]` 改为 `[x]`，并在末尾追加完成日期（`YYYY-MM-DD`）
- 编号规则：`<EPIC>-<序号>`，例如 `TG-001`
- 每日站会更新：昨天完成、今天计划、阻塞项

## 1. EPIC-BOOT（工程初始化）

- [x] `BOOT-001` 初始化 Go Module 与目录骨架（`cmd/`、`internal/`、`pkg/`、`migrations/`、`configs/`） - 2026-02-12
- [x] `BOOT-002` 新增 `Makefile` 基础命令：`run`、`test`、`lint`、`migrate-up`、`migrate-down` - 2026-02-12
- [x] `BOOT-003` 新增环境变量加载模块（支持 `.env` + 系统环境变量） - 2026-02-12
- [x] `BOOT-004` 定义配置结构体并实现启动时配置校验（缺关键项直接失败） - 2026-02-12
- [x] `BOOT-005` 接入结构化日志组件（JSON 输出） - 2026-02-12
- [x] `BOOT-006` 统一 `trace_id` 中间件与上下文传递 - 2026-02-12
- [x] `BOOT-007` 实现 `GET /healthz` 与 `GET /readyz` - 2026-02-12
- [x] `BOOT-008` 新增 Dockerfile（开发环境） - 2026-02-12
- [x] `BOOT-009` 新增 `docker-compose.yml`（app + postgres） - 2026-02-12
- [x] `BOOT-010` README 增加本地启动步骤 - 2026-02-12

## 2. EPIC-DB（数据库与迁移）

- [x] `DB-001` 创建 `users` 表 migration - 2026-02-12
- [x] `DB-002` 创建 `goals` 表 migration - 2026-02-12
- [x] `DB-003` 创建 `goal_profiles` 表 migration - 2026-02-12
- [x] `DB-004` 创建 `planning_sessions` 表 migration - 2026-02-12
- [x] `DB-005` 创建 `conversation_turns` 表 migration - 2026-02-12
- [x] `DB-006` 创建 `agent_action_logs` 表 migration - 2026-02-12
- [x] `DB-007` 创建 `message_dedup` 表 migration - 2026-02-12
- [x] `DB-008` 创建 `bot_runtime_states` 表 migration - 2026-02-12
- [x] `DB-009` 增加 `goals(user_id, status)` 索引 - 2026-02-12
- [x] `DB-010` 增加 `goal_profiles(goal_id, version_no desc)` 索引 - 2026-02-12
- [x] `DB-011` 增加 `planning_sessions(goal_id)` 唯一索引 - 2026-02-12
- [x] `DB-012` 增加 `conversation_turns(session_id, created_at)` 索引 - 2026-02-12
- [x] `DB-013` 编写初始种子数据脚本（可选） - 2026-02-12
- [x] `DB-014` 封装 DB 连接池参数（最大连接、空闲连接、生命周期） - 2026-02-12
- [x] `DB-015` 迁移脚本接入 CI 校验（up/down 可执行） - 2026-02-12

## 3. EPIC-TG（Telegram Long Polling）

- [ ] `TG-001` 实现 Telegram API Client：`getMe`
- [ ] `TG-002` 实现 Telegram API Client：`getUpdates`
- [ ] `TG-003` 实现 Telegram API Client：`sendMessage`
- [ ] `TG-004` `getUpdates` 支持 `timeout`、`offset`、`allowed_updates`
- [ ] `TG-005` 设计轮询循环（空结果继续轮询）
- [ ] `TG-006` 轮询失败重试（指数退避）
- [ ] `TG-007` 启动时从 `bot_runtime_states` 读取 `last_update_id`
- [ ] `TG-008` 每次处理完成后持久化最新 `last_update_id`
- [ ] `TG-009` 实现 `message_dedup` 写入与重复消息跳过
- [ ] `TG-010` 提取 Telegram Update -> 内部 Message DTO
- [ ] `TG-011` 仅允许处理 `message` 类型更新
- [ ] `TG-012` 过滤非文本消息并返回提示文案
- [ ] `TG-013` 实现命令解析器（`/start`、`/goal`、`/help`）
- [ ] `TG-014` 支持自然语言消息入口
- [ ] `TG-015` sender 增加失败重试（短重试）
- [ ] `TG-016` sender 429 限流处理（等待后重发）
- [ ] `TG-017` 轮询线程优雅退出（SIGTERM）
- [ ] `TG-018` 轮询状态指标上报（成功/失败次数）
- [ ] `TG-019` 轮询日志中打印 `update_id/chat_id`（脱敏）
- [ ] `TG-020` 进程重启恢复场景回归测试

## 4. EPIC-USER（用户与目标基础能力）

- [ ] `USER-001` 实现 `find_or_create_user_by_chat_id`
- [ ] `USER-002` 默认用户语言设为 `zh-CN`
- [ ] `USER-003` 默认用户时区设为 `Asia/Shanghai`
- [ ] `USER-004` 实现 `/start` 初始化欢迎流程
- [ ] `USER-005` 实现 `create_goal_draft`
- [ ] `USER-006` 实现 `get_active_goal_by_user_id`
- [ ] `USER-007` 无活跃目标时自动创建草稿目标
- [ ] `USER-008` 新建目标后写入 `goal_started` 埋点

## 5. EPIC-CLARIFY（澄清状态机）

- [ ] `CL-001` 定义状态枚举：`idle/clarifying/review/confirmed`
- [ ] `CL-002` 实现 `get_or_create_planning_session`
- [ ] `CL-003` session 每轮 `turn_count +1`
- [ ] `CL-004` 保存每轮用户输入到 `conversation_turns`
- [ ] `CL-005` 保存每轮 assistant 输出到 `conversation_turns`
- [ ] `CL-006` 实现意图路由器（`clarify_goal/confirm_plan/fallback_unknown`）
- [ ] `CL-007` `/goal` 进入 `clarifying` 状态
- [ ] `CL-008` 当必填字段补齐后进入 `review`
- [ ] `CL-009` 用户确认后进入 `confirmed`
- [ ] `CL-010` review 状态收到修改意见后回到 `clarifying`
- [ ] `CL-011` 每轮最多生成 1-2 个补问
- [ ] `CL-012` 每 3 轮自动触发“当前摘要 + 是否继续优化”
- [ ] `CL-013` 会话超时策略（例如 24h 无消息重置提醒）
- [ ] `CL-014` fallback 意图输出引导语并保留上下文

## 6. EPIC-SLOT（槽位抽取与规则校验）

- [ ] `SL-001` 定义 `GoalBriefDraft` 内部结构体
- [ ] `SL-002` 实现 `main_goal` 抽取与归一化
- [ ] `SL-003` 实现 `success_criteria` 抽取（数组）
- [ ] `SL-004` 校验 `success_criteria` 条数（3-5）
- [ ] `SL-005` 实现 `current_level` 抽取
- [ ] `SL-006` 实现 `time_budget.hours_per_week` 解析
- [ ] `SL-007` 实现 `time_budget.time_slots` 解析
- [ ] `SL-008` 实现 `constraints` 抽取
- [ ] `SL-009` 实现 `deadline` 识别（可空）
- [ ] `SL-010` 实现 `preferences` 抽取（可空）
- [ ] `SL-011` 实现 `risk_flags` 生成规则
- [ ] `SL-012` 实现“碎片化学习用户”识别规则
- [ ] `SL-013` 碎片化用户自动添加 15/30 分钟策略提示
- [ ] `SL-014` 截止时间与可投入冲突检测
- [ ] `SL-015` 约束与时间预算矛盾检测
- [ ] `SL-016` 非可验收成功标准检测与补问
- [ ] `SL-017` 输出 `open_questions`
- [ ] `SL-018` 计算 `completeness_score`
- [ ] `SL-019` 当必填缺失时禁止进入 review
- [ ] `SL-020` 冲突检测触发降级建议（降范围/延周期/降难度）

## 7. EPIC-TEMPLATE（模板引擎与渲染）

- [ ] `TPL-001` 新建 `schemas/goal_brief_v1.json`
- [ ] `TPL-002` 新建 `schemas/plan_pack_v1.json`（冻结，不启用生成）
- [ ] `TPL-003` 实现 Template Registry（按 `template_version` 注册）
- [ ] `TPL-004` 集成 JSON Schema Validator
- [ ] `TPL-005` 业务规则校验器（非 schema）
- [ ] `TPL-006` 校验失败输出 `error_code/field_path/repair_hint`
- [ ] `TPL-007` 组装最终 `Goal Brief v1` 结构化 JSON
- [ ] `TPL-008` 渲染 `Goal Brief` Markdown 文本
- [ ] `TPL-009` 确保 JSON 与 Markdown 同快照版本写库
- [ ] `TPL-010` 实现 `goal_profiles.version_no` 递增
- [ ] `TPL-011` `confirmation_state` 默认 `pending`
- [ ] `TPL-012` 用户确认后更新为 `confirmed`
- [ ] `TPL-013` 模板校验失败写入 `template_validation_failed` 埋点

## 8. EPIC-LLM（模型接入）

- [ ] `LLM-001` 抽象 `LLMProvider` 接口
- [ ] `LLM-002` 实现请求/响应 DTO
- [ ] `LLM-003` 构建槽位抽取 Prompt 模板
- [ ] `LLM-004` 构建补问生成 Prompt 模板
- [ ] `LLM-005` 强制 JSON 输出解析器
- [ ] `LLM-006` JSON 解析失败 fallback（规则补问）
- [ ] `LLM-007` 调用超时控制（10s）
- [ ] `LLM-008` 单次重试（指数退避）
- [ ] `LLM-009` 连续 2 次失败兜底话术
- [ ] `LLM-010` LLM 失败写 `agent_action_logs`
- [ ] `LLM-011` Prompt/Response 脱敏日志

## 9. EPIC-OBS（可观测与稳定性）

- [ ] `OBS-001` 增加埋点：`polling_cycle_started`
- [ ] `OBS-002` 增加埋点：`polling_cycle_succeeded`
- [ ] `OBS-003` 增加埋点：`polling_cycle_failed`
- [ ] `OBS-004` 增加埋点：`goal_started`
- [ ] `OBS-005` 增加埋点：`goal_brief_generated`
- [ ] `OBS-006` 增加埋点：`goal_brief_confirmed`
- [ ] `OBS-007` 增加埋点：`template_validation_failed`
- [ ] `OBS-008` 指标：`polling_cycle_success_rate`
- [ ] `OBS-009` 指标：`polling_update_lag_seconds_p95`
- [ ] `OBS-010` 指标：`goal_brief_generation_p95`
- [ ] `OBS-011` 统一错误码并输出到日志
- [ ] `OBS-012` 新增告警阈值配置（连续失败、延迟过高）

## 10. EPIC-TEST（测试与验收）

- [ ] `TEST-001` 单测：time budget 解析
- [ ] `TEST-002` 单测：success criteria 数量校验
- [ ] `TEST-003` 单测：deadline 冲突检测
- [ ] `TEST-004` 单测：completeness_score 计算
- [ ] `TEST-005` 单测：intent router 分流
- [ ] `TEST-006` 单测：state transition 合法性
- [ ] `TEST-007` 单测：schema validator
- [ ] `TEST-008` 单测：poll offset 恢复
- [ ] `TEST-009` 集成：polling 拉取 + 去重 + 入库
- [ ] `TEST-010` 集成：LLM 超时 fallback
- [ ] `TEST-011` 集成：用户确认链路 `pending -> confirmed`
- [ ] `TEST-012` 集成：sender 429 限流恢复
- [ ] `TEST-013` E2E：新用户 8 分钟完成 Goal Brief
- [ ] `TEST-014` E2E：碎片化用户自动策略
- [ ] `TEST-015` E2E：重启恢复不重复回复
- [ ] `TEST-016` 回归：命令入口与自然语言入口一致
- [ ] `TEST-017` 压测：并发 chat 处理稳定性
- [ ] `TEST-018` 压测：polling 周期稳定性

## 11. EPIC-RELEASE（发布准备）

- [ ] `REL-001` 完成 M1 配置清单文档（必填 env）
- [ ] `REL-002` 完成运行手册（启动/停止/回滚）
- [ ] `REL-003` 完成故障排查手册（轮询失败、DB 连接、LLM 超时）
- [ ] `REL-004` 完成验收演练脚本（产品 + 研发 + QA）
- [ ] `REL-005` 里程碑评审与缺陷清零
- [ ] `REL-006` 打 `m1` 发布 tag

## 12. 关键路径（必须按顺序）

1. `BOOT-*` + `DB-*` 完成  
2. `TG-001 ~ TG-010` 完成（可收发、可去重、可恢复）  
3. `CL-*` + `SL-*` + `TPL-*` 完成（可生成 Goal Brief）  
4. `LLM-*` + `OBS-*` 完成（可用且可观测）  
5. `TEST-*` 通过后进入 `REL-*`
