# Aiden PRD

## 0. 文档说明

- **文档目标**：定义 Aiden MVP 的产品边界、交付范围、验收标准与里程碑，确保 6 周内可上线验证。
- **当前版本**：v3.1（新增计划模板、目标澄清与首版计划输出契约、数据模型重构）。
- **适用阶段**：MVP（0 -> 1）。

## 1. 产品定义

**产品名称**：Aiden  
**一句话定位**：面向学生、求职者与碎片化学习者的对话式自学助手，帮助用户把学习方向拆成可执行计划，并通过打卡、手动调整与副目标待办池持续推进。  
**交互形态**：Telegram Bot（自然语言优先，命令为快捷入口）  
**技术偏好**：Go + PostgreSQL + LLM  
**MVP 主闭环**：目标澄清 -> 结构化目标摘要（Goal Brief）-> 首版动态计划（Plan Pack）-> 打卡 -> 手动调整 -> 副目标待办池管理

### 1.1 核心价值

- 帮用户从“模糊方向”快速进入“可执行状态”。
- 用滚动式计划降低一次性规划成本，支持长期学习。
- 用轻量反馈（打卡 + 手动调整）维持执行连续性。
- 用“主目标 + 副目标待办池”承接碎片化时间，不打断主线学习。
- 用统一计划模板保证可追溯、可调整、可迭代。

### 1.2 核心原则

- **自然语言优先**：用户不需要记复杂命令。
- **用户满意闸门**：计划是否收敛由用户决定，不设固定轮次结束。
- **动态计划机制**：计划总时长不设上限，系统按执行窗口滚动展开。
- **轻量多目标**：MVP 支持 1 个主目标 + 有上限的副目标待办池，不做完整并行排程。
- **模板先行**：目标澄清与首版计划必须产出可落入模板的结构化结果。
- **先闭环后扩展**：MVP 先做可用主链路，不做高复杂度自动化能力。

## 2. ICP 与使用场景

### 2.1 目标用户（ICP）

- **在校学生**：有实习/校招目标，需要阶段化学习推进。
- **求职者/转岗者**：有岗位方向，需要系统补课和项目准备。
- **碎片化学习者**：时间不连续，但愿意持续推进主目标并利用零散时间学习副主题。

### 2.2 入选标准（MVP 必须）

- 至少有 1 个明确主学习方向（可有或无硬截止日期）。
- 学习投入满足其一：
  - 每周可投入学习时间 >= 2 小时；
  - 或每周可提供 >= 5 个 15-30 分钟时间片。
- 愿意每周至少打卡 2 次（标准打卡或快速打卡）。
- 愿意通过对话调整计划与任务优先级。

### 2.3 暂不覆盖

- 同时管理 3 个以上同优先级主目标的重度并行用户。
- 纯兴趣探索且完全不关心阶段成果的用户。
- 团队/班级协作学习场景。

### 2.4 关键 JTBD

- 当我有学习方向但不知道怎么开始时，帮我拆出近期可执行任务。
- 当我只有 15-30 分钟碎片时间时，给我一个可以立即完成的小任务。
- 当我执行中掉队或卡住时，帮我快速调整计划而不是推倒重来。
- 当我不确定是否继续优化计划时，给我当前版本摘要，让我做决定。

## 3. 产品目标与边界（MVP）

### 3.1 MVP 目标（上线后前 4 周）

- 用户在 **8 分钟内**完成目标澄清并拿到首版计划。
- 目标澄清阶段输出标准化 `Goal Brief v1`（结构化字段完整）。
- 首版计划阶段输出标准化 `Plan Pack v1`（可直接落库、可渲染展示）。
- 用户可完成标准打卡或快速打卡，并查看周执行状态。
- 用户可通过自然语言完成手动调整并获得版本化结果。
- 用户可维护“副目标待办池”（新增/查看/归档/升级到主计划）。

### 3.2 MVP 非目标（后置）

- 自动调整计划（系统自动改计划）后置到 P1。
- 高级报告模板（社媒版/求职展示版）后置到 P1。
- 完整多目标并行排程与跨目标自动优先级后置到 P2。
- Web 管理后台、复杂可视化仪表盘不进入本期。
- 课程内容平台、社区能力不进入本期。

## 4. 端到端用户流程（MVP）

| 阶段 | 用户行为 | 系统行为 | 输出 | 继续条件 |
|---|---|---|---|---|
| A. 目标澄清 | 描述主学习方向与约束（可补充副兴趣） | 追问关键信息（每轮 1-2 个问题）并结构化槽位 | `Goal Brief v1` | 用户确认继续生成计划 |
| B. 首版计划 | 查看并反馈计划草案 | 将 Goal Brief 编译为计划模板实例 | `Plan Pack v1` | 用户确认开始执行 |
| C. 执行打卡 | 提交完成率/困难点/信心（或快速打卡） | 记录进度并汇总周状态 | Checkin 记录 + 周摘要 | 用户继续执行或请求调整 |
| D. 手动调整 | 用自然语言提出修改 | 识别意图、校验风险、生成新版本 | `Plan Pack vN` + 变更摘要 | 用户确认继续执行或继续优化 |
| E. 副目标池管理 | 新增/归档副目标，或将副目标升级到主计划 | 维护副目标池并给出影响提示 | Side Goal List + 升级建议 | 用户继续执行主目标或升级副目标 |

### 4.1 流程规则

- 计划优化循环无固定轮次上限。
- 每 3 轮对话系统自动输出一次“当前计划摘要 + 是否继续优化”。
- 用户可随时声明“当前计划满意，先执行”。
- 副目标默认不影响主目标节奏；仅在用户确认升级后进入主计划窗口。
- MVP 副目标池默认最多保持 3 个 Active 副目标，避免认知过载。

## 5. 功能需求（优先级 + 验收标准）

### 5.1 P0（MVP 必须）

#### F1. 目标澄清（Goal Brief 生成）

- **功能**：采集最小必要信息并检查冲突，生成 `Goal Brief v1`。
- **必要字段**：主学习目标、当前水平、可投入时间（总时长或时间片）、偏好/约束。
- **可选字段**：截止日期、副目标兴趣方向。
- **验收标准**：
  - 必填字段完整率 100%。
  - 截止日期缺失时自动进入“滚动学习模式”。
  - 若目标与时间冲突，必须给出降级建议（降范围/延周期/降难度）。
  - 若用户为碎片化投入，系统必须输出默认时间片策略（15/30 分钟）。
  - 输出必须通过 `Goal Brief v1` 模板校验。

#### F2. 首版动态计划生成（Plan Pack 生成）

- **功能**：基于 `Goal Brief v1` 生成 `Plan Pack v1`。
- **计划结构**：
  - 执行窗口（默认未来 2-4 周，可按用户要求扩展）。
  - 阶段里程碑（中长期目标节点）。
  - 每项任务的预计时长与验收标准。
  - 碎片任务位（适配 15-30 分钟时间片的可完成动作）。
- **验收标准**：
  - 不设计划总周期上限。
  - 每个执行窗口至少包含 3 项任务（可学习、可实践、可复盘）。
  - 碎片化学习用户每周至少 2 项 <= 30 分钟任务。
  - `Plan Pack` 模板校验通过率 >= 98%。
  - 计划生成延迟 P95 <= 60 秒。

#### F3. 打卡与进度记录

- **功能**：记录执行反馈并展示周状态。
- **标准打卡输入字段**：完成率（0-100）、困难点、信心（1-5）、可选备注。
- **快速打卡输入字段**：是否完成（是/否）+ 一句话阻塞说明（可选）。
- **验收标准**：
  - 打卡写入成功率 >= 99%。
  - 支持按周汇总：已完成、未完成、风险项（含主/副目标维度）。
  - 支持用户补打卡（过去 7 天）。
  - 快速打卡流程在 20 秒内可完成（可用性验收）。

#### F4. 手动计划调整（PlanOps）

- **功能**：响应用户的计划修改请求并落库新版本。
- **动作**：`adjust_plan`（含任务重排、节奏调整、副目标升级）。
- **验收标准**：
  - 用户可用自然语言触发，无需命令。
  - 高影响变更（跨周重排或任务改动 >30%）必须二次确认。
  - 返回“修改前后对比 + 修改原因 + 影响范围”。
  - 支持“副目标 -> 主计划任务”的显式升级流程。
  - 调整响应延迟 P95 <= 30 秒。

#### F5. 计划版本管理与可追溯

- **功能**：保存每次主计划变更历史。
- **验收标准**：
  - 每次主计划调整生成新版本号（v1, v2...）。
  - 可查看最近 5 个版本摘要。
  - 支持回滚至上一个版本。
  - 副目标池操作（新增/归档/升级）需具备可追溯日志。

#### F6. 交互与命令

- **命令（可选）**：`/start` `/goal` `/plan` `/checkin` `/adjust` `/sidegoal` `/help`
- **验收标准**：
  - 全流程可仅靠自然语言完成。
  - 命令与自然语言入口行为一致（同意图同结果）。

#### F7. 轻量多目标机制（主目标 + 副目标待办池）

- **功能**：
  - 支持 1 个主目标 + 最多 3 个 Active 副目标。
  - 副目标字段包含：标题、状态、优先级、next action、预计分钟数。
  - 支持新增、查看、归档、完成、升级到主计划。
- **验收标准**：
  - 新增副目标最多 3 轮对话内完成。
  - 查看副目标池时必须展示 next action 与预计耗时。
  - 升级副目标时必须展示对主计划影响并二次确认。
  - 未升级前，副目标不自动改写当前主计划。

#### F8. 模板引擎与双层输出

- **功能**：统一模板版本、字段校验、渲染输出，支持结构化与可读文本双层结果。
- **验收标准**：
  - 所有 `Goal Brief` 与 `Plan Pack` 均带 `template_version`。
  - 结构化存储（JSON）与用户可读渲染（Markdown）保持同一版本快照。
  - 模板校验失败时返回可解释错误并触发补问，不直接生成计划。

### 5.2 P1（后置）

- 自动调整（基于打卡自动提出并应用计划变更）。
- 学习报告能力与高级模板（社媒版/求职版）。
- 打卡提醒/复盘提醒。
- 来源可信度自动评分与来源分级展示。
- 双主目标并行（有限场景，非自动排程）。

### 5.3 P2（增强）

- 完整多目标并行管理（多主目标 + 跨目标优先级）。
- Markdown/PDF 导出。
- 英文与多语言支持。

## 6. 对话与决策策略

### 6.1 对话策略

- 每轮仅问 1-2 个最关键问题。
- 先确认约束，再给建议；避免无约束生成计划。
- 对碎片化用户优先提供“15-30 分钟即可完成”的下一步动作。
- 每次调整后必须询问：`当前计划你是否满意，还是继续优化？`

### 6.2 意图路由（MVP）

- `clarify_goal`：补充目标信息并填充 Goal Brief。
- `generate_plan`：根据 Goal Brief 生成/刷新 Plan Pack。
- `checkin_progress`：写入标准打卡。
- `quick_checkin`：写入快速打卡。
- `adjust_plan`：修改计划结构与节奏。
- `manage_side_goal`：新增/归档/升级副目标。
- `confirm_plan`：用户确认进入执行。

### 6.3 模板填充策略

- 先填必填槽位，再填优化槽位（偏好、风格、资源）。
- 缺少关键槽位时只能补问，不能直接跳过。
- 允许“弹性扩展字段”写入 `custom_blocks`，避免过度硬编码。
- 每次生成后返回“模板完整度”与“待确认项”。

### 6.4 异常与兜底

- LLM 超时/失败：返回模板化保守建议 + 引导重试。
- 用户输入过少：追问单一关键缺口，不并发多问题。
- 连续两次完成率 < 50%：建议发起手动调整。
- Active 副目标 > 3：提示先归档低优先级副目标，再新增。

## 7. 计划模板与输出契约（新增）

### 7.1 模板设计目标

- 既能结构化落库与计算，又能自然语言展示。
- 对不同学习方向（编程/语言/考试）保持通用。
- 在 MVP 内只冻结“核心骨架”，允许扩展字段灵活生长。

### 7.2 Goal Brief v1 模板（目标澄清输出）

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `template_version` | string | 是 | 固定 `goal_brief_v1` |
| `goal_id` | uuid | 是 | 主目标 ID |
| `main_goal` | string | 是 | 最终目标描述 |
| `success_criteria` | string[] | 是 | 3-5 条可验收标准 |
| `current_level` | string | 是 | 当前水平自评 |
| `time_budget` | object | 是 | `hours_per_week` 或 `time_slots` |
| `constraints` | string[] | 是 | 时间/设备/心理负担等 |
| `preferences` | string[] | 否 | 学习偏好 |
| `deadline` | date/null | 否 | 无则滚动学习 |
| `side_goal_candidates` | object[] | 否 | 候选副目标（最多 3） |
| `risk_flags` | string[] | 是 | 识别到的风险 |
| `confirmation_state` | enum | 是 | `pending` / `confirmed` |
| `custom_blocks` | jsonb | 否 | 弹性扩展字段 |

### 7.3 Plan Pack v1 模板（首版计划输出）

**顶层结构**

| 区块 | 必填 | 说明 |
|---|---|---|
| `plan_meta` | 是 | `plan_id`、`version`、`template_version`、`created_at` |
| `goal_snapshot` | 是 | Goal Brief 摘要快照 |
| `execution_strategy` | 是 | 周节奏、窗口策略、优先级策略 |
| `stages[]` | 是 | 3-6 个阶段，默认 4 阶段 |
| `weekly_rhythm` | 是 | 每周执行模板（微学习/深度块/复盘） |
| `side_goal_pool` | 是 | Active 副目标列表与 next action |
| `adjustment_triggers` | 是 | 触发调整的阈值与动作 |
| `checkin_contract` | 是 | 标准/快速打卡口径 |
| `references` | 否 | 资源或依据链接 |
| `custom_blocks` | 否 | 扩展区 |

**阶段结构 `stages[]`**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `stage_id` | string | 是 | 如 `S1`、`S2` |
| `name` | string | 是 | 阶段名称 |
| `duration_weeks` | int | 是 | 阶段持续周数 |
| `objective` | string | 是 | 阶段目标 |
| `deliverable` | string | 是 | 阶段可交付产物 |
| `entry_criteria` | string[] | 否 | 进入条件 |
| `exit_criteria` | string[] | 是 | 退出验收条件 |
| `tasks[]` | object[] | 是 | 主线任务列表 |
| `micro_tasks[]` | object[] | 是 | 15-30 分钟任务位 |
| `side_goal_focus` | string[] | 否 | 本阶段关注的副目标 |

**任务结构 `tasks[] / micro_tasks[]`**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `task_id` | string | 是 | 任务 ID |
| `title` | string | 是 | 任务名称 |
| `task_type` | enum | 是 | `learn/practice/review/project/bridge` |
| `est_minutes` | int | 是 | 预计时长 |
| `acceptance_criteria` | string | 是 | 完成标准 |
| `output_artifact` | string | 否 | 产出物（代码/笔记/录屏） |
| `depends_on` | string[] | 否 | 依赖任务 |
| `priority` | enum | 是 | `high/medium/low` |

### 7.4 模板灵活性机制

- 核心字段冻结，扩展字段统一进入 `custom_blocks`。
- 阶段数可在 3-6 之间动态变化，不固定 4 阶段。
- `tasks` 支持不同类型配比，不强制固定配方。
- 可按用户类型切换周节奏（小时制或时间片制）。

### 7.5 输出契约（Agent 必须遵守）

- **目标澄清结束输出**：`Goal Brief v1`（结构化 + 文本摘要）。
- **首版计划结束输出**：`Plan Pack v1`（结构化 + 文本计划）。
- 两步输出均需给出：
  - `template_version`
  - `completeness_score`（0-100）
  - `open_questions`（若有）

## 8. 数据模型（按计划结构重构）

### 8.1 核心实体

- `User`：用户基础信息、时区、语言、偏好。
- `Goal`：主目标主记录（状态、当前生效计划版本、生命周期）。
- `GoalProfile`：`Goal Brief` 结构化存储（对应目标澄清输出）。
- `PlanVersion`：计划版本元数据（`version_no`、来源、是否生效、模板版本）。
- `PlanDocument`：`Plan Pack` 完整 JSON 快照 + 文本渲染快照。
- `PlanStage`：阶段结构（名称、时长、目标、交付、退出条件）。
- `PlanTask`：任务结构（类型、时长、验收、优先级、依赖、是否微任务）。
- `SideGoal`：副目标池（标题、状态、优先级、next action、预计分钟数、升级记录）。
- `Checkin`：打卡记录（标准/快速、关联阶段/任务/副目标、信心、阻塞点）。
- `PlanChangeLog`：版本差异、影响范围、确认记录、回滚关系。
- `PlanningSession`：澄清会话上下文（轮次、槽位完成度）。
- `ConversationTurn`：每轮对话消息与意图置信度。
- `AgentActionLog`：动作执行日志（参数、状态、错误码、确认链路）。

### 8.2 关键关系与约束

- `Goal` 1:N `PlanVersion`，且仅 1 个 `active_plan_version_id`。
- `PlanVersion` 1:1 `PlanDocument`；`PlanVersion` 1:N `PlanStage`；`PlanStage` 1:N `PlanTask`。
- `Goal` 1:N `SideGoal`，并约束 `active_side_goal_count <= 3`。
- `Checkin` 可关联 `plan_task_id` 或 `side_goal_id`（二者至少其一）。
- 副目标升级必须事务化：`SideGoal.status=promoted` + 新建 `PlanTask(task_type=bridge)` + 写 `PlanChangeLog`。

### 8.3 模板到数据模型映射

| 模板区块 | 存储对象 |
|---|---|
| `Goal Brief v1` | `GoalProfile` |
| `plan_meta`/`execution_strategy`/`weekly_rhythm`/`adjustment_triggers` | `PlanDocument` + `PlanVersion` |
| `stages[]` | `PlanStage` |
| `tasks[]`/`micro_tasks[]` | `PlanTask` |
| `side_goal_pool` | `SideGoal` |
| `checkin_contract` 与执行数据 | `Checkin` |
| 版本差异与确认 | `PlanChangeLog` |

### 8.4 数据口径补充

- 所有计划相关事件必须携带：`goal_id` `plan_version` `template_version` `timestamp`。
- 打卡去重键：`user_id + goal_id + target_id(task/side_goal) + date + checkin_type`。
- 模板校验失败必须落 `AgentActionLog`，用于诊断补问质量。

## 9. 指标与埋点

### 9.1 北极星与核心指标（上线后前 4 周）

- **激活率**：完成“目标澄清 + 首版计划”占比 >= 60%。
- **计划确认率**：看到首版计划后确认执行占比 >= 70%。
- **D7 打卡留存**：注册后第 7 天至少 1 次打卡占比 >= 35%。
- **周执行率**：周任务平均完成率 >= 55%。
- **调整使用率**：有打卡用户中使用过手动调整占比 >= 30%。
- **碎片学习激活率**：有打卡用户中使用过“快速打卡或副目标池”占比 >= 40%。
- **模板通过率**：Goal Brief/Plan Pack 模板校验通过率 >= 98%。

### 9.2 核心事件

- `goal_started`
- `goal_brief_generated`
- `goal_brief_confirmed`
- `plan_pack_generated`
- `plan_confirmed`
- `checkin_submitted`
- `quick_checkin_submitted`
- `adjust_requested`
- `adjust_applied`
- `plan_rolled_back`
- `sidegoal_added`
- `sidegoal_promoted`
- `sidegoal_archived`
- `template_validation_failed`

## 10. 技术方案（MVP）

### 10.1 系统分层

- **交互层**：Telegram Bot API（Webhook）。
- **应用层**：对话状态机、意图路由器、计划引擎、PlanOps 执行器、副目标池管理器。
- **模板层**：模板注册中心（Schema Registry）、模板校验器（Validator）、渲染器（Renderer）。
- **数据层**：PostgreSQL（事务写入 + 版本化存储 + JSONB 快照）。
- **AI 层**：LLM（信息抽取、计划生成、调整说明文本化）。

### 10.2 非功能要求

- 可用性：核心链路成功率 >= 99%。
- 性能：计划生成 P95 <= 60 秒；调整 P95 <= 30 秒。
- 可追溯：所有主计划改动和副目标关键操作必须可回看。
- 一致性：副目标升级到主计划需在单事务内完成状态变更与日志写入。
- 兼容性：模板小版本升级不破坏历史计划读取。
- 安全：仅存储必要用户信息；敏感数据脱敏日志化。

## 11. 里程碑与交付（6 周）

### M1（第 1-2 周）：目标澄清与模板冻结

- 完成 Telegram 接入、用户/主目标存储。
- 完成澄清状态机、字段校验、冲突检测、时间片识别。
- 冻结 `Goal Brief v1` 与 `Plan Pack v1` 模板。
- 打通“目标澄清 -> Goal Brief”链路。
- **出门标准**：用户可从 0 到 1 生成并确认 `Goal Brief v1`。

### M2（第 3-4 周）：首版计划、打卡与副目标池

- 上线 `Goal Brief -> Plan Pack v1` 生成链路。
- 上线标准打卡/快速打卡、周汇总、补打卡。
- 上线副目标待办池（新增/查看/归档/升级）。
- 完成数据模型重构（PlanDocument/PlanStage/PlanTask/SideGoal）。
- 完成核心埋点和基础看板。
- **出门标准**：用户可执行、可反馈、可追溯，并可管理副目标池。

### M3（第 5-6 周）：手动调整与稳定性

- 上线 `adjust_plan` 全链路与高影响二次确认。
- 打通副目标升级对主计划影响提示与确认。
- 完成性能压测、错误兜底、灰度发布。
- 完成 20-50 名种子用户闭环验证。
- **出门标准**：闭环稳定运行，关键指标达到基线。

## 12. 风险与应对

- **目标描述模糊，计划质量不稳定**  
  应对：强制最小字段 + 冲突检测 + 模板化降级建议。

- **碎片任务过多导致“忙但不前进”**  
  应对：主目标任务优先级高于副目标；周摘要明确“主目标推进度”。

- **副目标池膨胀导致认知负担**  
  应对：Active 副目标上限 3；超限时引导归档或升级。

- **模板过硬导致场景不适配**  
  应对：冻结核心字段，开放 `custom_blocks` 扩展。

- **LLM 输出波动**  
  应对：结构化输出约束 + 字段级校验 + 失败回退模板。

- **范围失控导致延期**  
  应对：严格 P0/P1/P2 边界；自动调整和完整并行多目标不进本期。

## 13. 示例：Python 入门（按模板填充）

### 13.1 Goal Brief v1 示例

- `main_goal`：8 周内完成 Python 入门，并独立交付 1 个可运行命令行项目（任务管理器 CLI）。
- `success_criteria`：
  - 完成 40+ 道基础练习题；
  - 完成 300-500 行项目代码；
  - 项目具备 10+ 个基础 `pytest` 用例；
  - 项目通过 `ruff check` 与 `black --check`；
  - 提交 GitHub README（运行方式+功能说明）。
- `current_level`：零基础，了解变量概念但无项目经验。
- `time_budget`：每周 `5 x 25min` + `1 x 90min`，总计约 3.5 小时。
- `constraints`：工作日晚间时间碎片化，周中注意力有限。
- `side_goal_candidates`：
  - Git 最小上手；
  - 英文报错关键词；
  - 终端效率。

### 13.2 Plan Pack v1 示例（4 阶段）

**阶段 S1（第 1-2 周）语法与运行环境**

- `objective`：能独立写并运行基础脚本。
- `deliverable`：命令行计算器 v1。
- `tasks`：
  - 基础语法与流程控制（120 分钟，`learn`）；
  - 10-12 道练习（180 分钟，`practice`）；
  - 函数重构练习（60 分钟，`review`）。
- `micro_tasks`：
  - 完成 1 道 if/loop 小题（20 分钟）；
  - 修复 1 个语法报错并记录原因（15 分钟）。
- `side_goal_focus`：Git 最小上手。

**阶段 S2（第 3-4 周）数据结构与模块化**

- `objective`：掌握列表/字典/集合、文件读写、模块拆分。
- `deliverable`：通讯录 CLI（含 JSON 持久化）。
- `tasks`：
  - 数据结构增删改查（120 分钟，`learn`）；
  - 文件/JSON 持久化（120 分钟，`practice`）；
  - 模块拆分（90 分钟，`project`）。
- `micro_tasks`：
  - 写 1 个列表推导式练习（15 分钟）；
  - 用 `pathlib` 完成路径处理（20 分钟）。
- `side_goal_focus`：调试能力、标准库探索。

**阶段 S3（第 5-6 周）稳健性与工程习惯**

- `objective`：具备异常处理、基础测试和代码规范能力。
- `deliverable`：任务管理器 CLI（核心功能 + 测试）。
- `tasks`：
  - 异常处理与日志（120 分钟，`learn`）；
  - `pytest` 用例编写（150 分钟，`practice`）；
  - `ruff` + `black` 接入（60 分钟，`review`）。
- `micro_tasks`：
  - 增加 1 个边界测试（20 分钟）；
  - 修复 1 个 lint 警告（15 分钟）。
- `side_goal_focus`：测试覆盖、代码重构。

**阶段 S4（第 7-8 周）交付与展示**

- `objective`：完成可展示作品并形成复盘闭环。
- `deliverable`：可公开展示的 Python 入门项目。
- `tasks`：
  - 完善 README 与使用说明（90 分钟，`project`）；
  - 录制 3 分钟演示（60 分钟，`project`）；
  - 技术复盘与下一阶段路线（60 分钟，`review`）。
- `micro_tasks`：
  - 补 1 条已知问题说明（15 分钟）；
  - 准备 1 段项目讲解稿（20 分钟）。
- `side_goal_focus`：项目表达、下一阶段路线。

**副目标池（全程）**

- Active 上限 3；默认不打断主线。
- 升级条件：连续 2 次主线卡住、可直接解除阻塞、用户显式确认其一满足。

### 13.3 示例依据与置信说明

- **高置信（官方技术路径）**：
  - Python 官方教程目录与学习顺序：
    - https://docs.python.org/3/tutorial/index.html
    - https://docs.python.org/3/tutorial/controlflow.html
    - https://docs.python.org/3/tutorial/datastructures.html
    - https://docs.python.org/3/tutorial/modules.html
    - https://docs.python.org/3/tutorial/errors.html
    - https://docs.python.org/3/tutorial/classes.html
    - https://docs.python.org/3/tutorial/venv.html
- **高置信（工程工具实践）**：
  - https://docs.pytest.org/en/stable/getting-started.html
  - https://docs.astral.sh/ruff/
  - https://black.readthedocs.io/en/stable/the_black_code_style/current_style.html
  - https://peps.python.org/pep-0008/
- **高置信（学习科学证据）**：
  - Cepeda et al., 2006（分散学习效应）：https://pubmed.ncbi.nlm.nih.gov/16719566/
  - Karpicke & Blunt, 2011（检索练习）：https://pubmed.ncbi.nlm.nih.gov/21252317/
  - Dunlosky et al., 2013（学习策略综述）：https://pubmed.ncbi.nlm.nih.gov/26173288/

## 14. 版本冻结决策（M1 前必须冻结）

- MVP 仅支持中文。
- MVP 支持 **1 个主目标 + 最多 3 个 Active 副目标**。
- MVP 必须输出 `Goal Brief v1` 与 `Plan Pack v1`（模板版本冻结）。
- MVP 不做自动调整，只做手动调整。
- MVP 不做完整多目标并行排程。
- MVP 不做高级报告模板与社媒文案能力。
