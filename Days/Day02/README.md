# 主题

类型系统、零值与领域建模。

今天把 `GoTaskFlow` 从 Day 1 的工程骨架推进到第一版核心领域模型。课程条目要求覆盖 `struct`、method、value receiver、pointer receiver、常量枚举、时间建模和领域不变量；项目实践是设计 `Job`、`Queue`、`Worker`、`JobStatus` 等核心模型。重点不是把字段堆满，而是让类型先表达业务规则，减少无效状态进入系统的机会。

# 学习重点

Go 的 `struct` 是领域数据的主要承载方式。字段命名、字段类型和构造函数共同定义了模型边界。比如任务 ID 可以先用 `type JobID string` 与普通字符串区分，任务状态可以用 `type JobStatus string` 加常量集合表达有限状态。

method 让行为贴近数据。值接收者适合只读行为，例如 `Summary`、`Age`、`CanRetry`；指针接收者适合会修改对象状态的行为，例如 `Start`、`Complete`、`Fail`、`Retry`。接收者选择不是语法偏好，而是模型是否会发生状态变化的声明。

零值是 Go 类型系统的重要设计约束。一个类型的零值要么可直接使用，要么必须能被清晰识别为无效值。今天的 `Queue` 零值可以直接 `Enqueue`，因为内部 slice 的零值可追加；`Job` 零值则不是合法任务，因此提供 `IsZero` 和 `NewJob` 来隔离无效输入。

时间建模不能只用字符串。`time.Time` 能表达创建时间、开始时间、完成时间，并能通过 `IsZero` 区分“尚未发生”。这让任务生命周期可以被测试、排序和计算耗时。

领域不变量要尽早进入模型。空标题、空 ID、负重试次数、非法状态流转都不应该等到 CLI 或存储层才发现。核心模型越早拒绝无效状态，后续仓储、调度、HTTP API 和测试就越稳定。

# 项目实践

今天新增 `Days/Day02/Day02_domain_model.go`，用单文件形式建立 GoTaskFlow 的领域模型原型：

```text
JobID / WorkerID        领域 ID 类型
JobStatus               有限状态枚举
JobSpec                 创建任务时的输入契约
Job                     任务实体与生命周期行为
Worker                  执行任务的工作者模型
Queue                   任务队列，零值可用
```

代码会演示三类关键行为：

1. `NewJob` 与 `NewWorker` 在入口处校验领域不变量。
2. `Job` 的方法维护任务状态流转。
3. `Queue` 通过值接收者返回副本，避免调用方直接修改内部 slice。

# 核心理解

Go 的类型系统强调组合、显式状态和零值可用；好的领域模型先约束错误状态。

领域建模的目标不是做一个“看起来像业务名词”的结构体集合，而是让代码中的状态变化必须经过受控路径。只要 `Job` 的状态只能通过 `Enqueue`、`Start`、`Complete`、`Fail`、`Cancel`、`Retry` 等方法变化，后续服务层就不需要在每个入口重复猜测当前状态是否合法。

# 参考源

- https://go.dev/ref/spec#Struct_types
- https://go.dev/ref/spec#Method_declarations
- https://go.dev/doc/effective_go#methods

# 今日目标

1. 理解 `struct` 如何承载领域数据和边界。
2. 理解 value receiver 与 pointer receiver 的工程含义。
3. 用常量枚举表达 `JobStatus` 的有限状态集合。
4. 用 `time.Time` 表达任务生命周期中的关键时间点。
5. 为 `Job`、`Queue`、`Worker` 建立第一版领域不变量。
6. 运行 Day 2 代码，观察合法状态流转和非法状态流转的差异。

# 知识展开

Go 的结构体类型由字段列表定义。字段的类型越具体，调用方越难把无关数据传进来。`JobID` 和 `WorkerID` 底层都是 `string`，但独立命名后，方法签名会更清楚：`Start(worker Worker, now time.Time)` 比 `Start(workerID string)` 能携带更多业务语义，也给后续扩展 Worker 能力留出空间。

常量枚举在 Go 中通常不是 `enum` 关键字，而是命名类型加一组常量。今天的状态包括 `pending`、`queued`、`running`、`succeeded`、`failed`、`canceled`。`Valid` 方法负责判断状态是否在允许集合中，`canTransition` 负责描述状态机规则。

值接收者会复制接收者值，适合不改变状态的方法。`Job.Summary` 和 `Job.Age` 只读取字段，所以使用值接收者可以表达“这个方法不会修改任务”。`Queue.Jobs` 同样使用值接收者，并返回内部 slice 的副本，避免调用方拿到内部存储。

指针接收者适合修改状态。`Job.Start` 会设置 `Status`、`StartedAt`、`WorkerID` 和 `Attempts`，必须使用指针接收者，否则修改只发生在副本上。这个选择会直接影响代码是否符合预期。

零值可用要分类型判断。`Queue{}` 可以作为空队列使用，因为 `nil` slice 可以 append；但 `Job{}` 没有 ID、标题和状态，不应该被当作业务任务。因此 `Queue` 的零值是可用状态，`Job` 的零值是可检测的非法状态。

时间字段建议保留为 `time.Time`，而不是提前格式化成字符串。显示时可以格式化，存储时可以序列化，但领域层应保留可比较、可计算、可校验的时间类型。

# 代码示例

运行默认演示：

```bash
go run Days/Day02/Day02_domain_model.go
```

查看初始队列：

```bash
go run Days/Day02/Day02_domain_model.go list
```

查看状态统计：

```bash
go run Days/Day02/Day02_domain_model.go counts
```

观察非法状态流转：

```bash
go run Days/Day02/Day02_domain_model.go invalid
```

`invalid` 命令会尝试把刚创建的 `pending` 任务直接 `Complete`。领域模型会拒绝这条路径，因为任务必须先入队、开始执行，才能完成。这是领域不变量在代码中的直接体现。

# 项目任务拆解

1. 定义 `JobID`、`WorkerID`，避免所有标识都退化成普通字符串。
2. 定义 `JobStatus` 常量集合和 `Valid` 校验方法。
3. 定义 `JobSpec` 作为创建任务的输入契约。
4. 实现 `NewJob`，校验 ID、标题、重试次数和默认队列。
5. 实现 `Job` 生命周期方法：入队、开始、完成、失败、取消、重试。
6. 定义 `Worker`，让任务执行归属有明确类型。
7. 定义零值可用的 `Queue`，支持入队、列表、按状态过滤和状态统计。
8. 提供 CLI 演示命令，让模型行为可以被 `go run` 验证。

# 常见误区

把所有字段都设为 `string`。这样初期写得快，但后续无法通过类型区分 ID、状态、队列名和普通文本。

用字段赋值绕过构造函数。如果调用方可以随意构造 `Job{Status: "done"}`，领域不变量就失效了。实际项目中应逐步把字段可见性收紧到包内。

误用指针接收者或值接收者。会修改状态的方法必须使用指针接收者；只读方法优先使用值接收者，让调用方更容易判断副作用。

忽略零值。零值不是边角问题，而是 Go 代码经常自然出现的状态。模型必须决定零值可用还是可检测为无效。

把时间当成展示字符串。领域层需要计算年龄、耗时、超时和排序，应该保留 `time.Time`。

把状态流转散落在调用方。状态机规则应集中在领域模型内，否则服务层、CLI、测试和未来 HTTP handler 会重复实现规则。

# 自测问题

1. 为什么 `Queue` 的零值可以直接使用，而 `Job` 的零值不应该作为合法任务？
2. `Job.Start` 为什么必须使用 pointer receiver？
3. `Job.Summary` 为什么适合使用 value receiver？
4. `JobStatus` 为什么需要 `Valid` 方法？
5. `pending -> succeeded` 为什么是不合法流转？
6. `time.Time.IsZero` 在任务生命周期建模中解决了什么问题？
7. `JobSpec` 和 `Job` 分开有什么好处？

# 验收标准

1. `Days/Day02/README.md` 包含所有要求章节，且内容基于 Day 2 syllabus 条目展开。
2. `Days/Day02/Day02_domain_model.go` 可以通过 `gofmt`。
3. `go run Days/Day02/Day02_domain_model.go` 可以成功执行默认演示。
4. `go run Days/Day02/Day02_domain_model.go invalid` 会返回非法状态流转错误。
5. 代码中能看到 `struct`、method、value receiver、pointer receiver、常量枚举、时间建模和领域不变量。
6. `Queue` 零值可用，`Job` 零值可检测为无效。

# 明日衔接

Day 3 将进入 Slice、Map 与内存语义。今天的 `Queue` 已经使用 slice 保存任务，并通过复制返回结果保护内部状态；明天会进一步实现内存版 Job Repository，深入处理 slice header、map 引用语义、容量扩展、集合建模和查询过滤。

# GitHub Commit

- 仓库：/Users/liushan/Documents/Codex/Golang-Learning
- 代码目录：`Days/Day02/`
- Commit：`待提交`
- Commit 链接：待提交
- Push 状态：待提交
