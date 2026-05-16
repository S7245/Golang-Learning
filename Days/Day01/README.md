# 主题

Go 工程化环境与模块边界。

今天的目标不是写一个零散脚本，而是为连续项目 `GoTaskFlow` 建立可演进的工程起点。`GoTaskFlow` 会从本地 CLI 任务队列与工作流引擎开始，后续逐步加入并发调度、持久化、HTTP 服务、测试、可观测性与架构治理。Day 1 重点放在 Go toolchain、module、workspace、`cmd/` 与 `internal/` 分层，以及可执行程序入口的职责边界。

# 学习重点

`go` 命令是 Go 工程的入口。安装工具链后，常用动作包括 `go run`、`go test`、`go build`、`go fmt`、`go mod init`、`go mod tidy`。这些命令共同定义了代码如何被发现、编译、测试和发布。

Go module 是依赖与导入路径的边界。`go.mod` 里的 module path 决定本仓库包被其他代码导入时的名字，也决定仓库内部导入路径的稳定性。一个项目一旦进入协作或发布阶段，module path 就不应随意变动。

`cmd/` 通常放可执行程序入口，例如 `cmd/gotaskflow/main.go`。入口层应该薄，主要负责解析参数、组装依赖、调用应用服务并把错误转换成用户可理解的输出。

`internal/` 是 Go 原生支持的可见性边界。放入 `internal/` 下的包只能被父目录树内的代码导入，适合承载项目私有领域模型、服务、仓储、调度器等实现。

workspace 适合多模块开发，但 Day 1 先把单模块边界打稳。后续如果项目拆出 SDK、示例、插件或独立工具，再考虑 `go work`。

# 项目实践

今天初始化 `GoTaskFlow` 的工程骨架思路：

```text
go.mod
cmd/
  gotaskflow/
    main.go
internal/
  taskflow/
    job.go
    queue.go
README.md
Days/
  Day01/
    README.md
    Day01_module_layout.go
```

本日代码文件 `Day01_module_layout.go` 是一个可直接运行的 Day 1 原型。它用单文件方式模拟未来 `cmd/` 和 `internal/` 的职责：CLI 层处理命令，领域层维护任务队列，布局清单描述未来包边界。

# 核心理解

Go 项目不是脚本集合，而是由模块路径、包边界、构建约束共同定义的工程系统。

工程化边界的核心是让变化有位置：命令行参数变化应停留在 `cmd/`，领域规则变化应停留在内部领域包，存储替换应停留在仓储实现，外部接口变化不应污染核心模型。

# 参考源

- https://go.dev/doc/install
- https://go.dev/ref/mod
- https://pkg.go.dev/cmd/go

# 今日目标

1. 理解 Go toolchain 在项目生命周期中的角色。
2. 理解 module path 与 import path 的关系。
3. 理解 `cmd/` 入口和 `internal/` 私有包的职责边界。
4. 运行 Day 1 代码，看到 `GoTaskFlow` 的最小 CLI 与布局清单。
5. 为后续 50 天连续演进保留稳定目录约定。

# 知识展开

安装 Go 后，先确认工具链可用：

```bash
go version
go env GOPATH GOMOD GOWORK
```

`GOMOD` 指向当前生效的 `go.mod`。如果为空，说明当前目录不在模块中。Day 1 的核心不是记住所有 `go env` 输出，而是知道当构建结果异常时，应该先检查当前模块、工作区和 Go 版本。

`go.mod` 是模块声明，不只是依赖清单。典型初始化命令如下：

```bash
go mod init github.com/S7245/Golang-Learning
go mod tidy
```

`go mod tidy` 会根据代码实际导入修正依赖。早期项目没有外部依赖时，它通常不会产生很多变化，但养成 tidy 的习惯可以避免依赖漂移。

`cmd/` 的价值在于把可执行入口与业务能力分开。未来 `gotaskflow` 可能有 `add`、`list`、`run`、`retry`、`inspect`、`stats` 等命令，但这些命令不应该直接持有所有业务规则。入口层应该把请求翻译成应用服务调用。

`internal/` 的价值是防止未承诺的实现被外部依赖。`internal/taskflow` 可以承载 `Job`、`Queue`、`Worker`、`Dispatcher` 等内部模型。只要它们还不是公共 API，就应该避免放到仓库根目录的公开包里。

Day 1 先采用单模块，因为连续项目的主要复杂度会来自领域能力、并发、持久化和服务边界。过早拆成多个 module 会增加版本、replace、workspace 的维护成本。

# 代码示例

运行今天的代码：

```bash
go run Days/Day01/Day01_module_layout.go
go run Days/Day01/Day01_module_layout.go add "write first task"
go run Days/Day01/Day01_module_layout.go list
```

第一个命令会打印建议的工程布局。`add` 和 `list` 展示 CLI 层如何调用一个最小任务队列。由于今天还没有持久化，任务只存在于本次进程内；这正好为后续的领域模型、仓储和文件持久化留下演进空间。

# 项目任务拆解

1. 建立 `Days/Day01/` 保存当天教程和代码。
2. 用单文件原型定义 `GoTaskFlow` 的模块布局清单。
3. 用最小 CLI 命令模拟未来 `cmd/gotaskflow` 的入口职责。
4. 用 `Job` 与 `Queue` 模拟未来 `internal/taskflow` 的核心领域位置。
5. 通过 `go run` 验证代码可以执行。

# 常见误区

把 `main.go` 写成全项目业务中心。入口文件应该薄，业务规则要移动到可测试、可复用的内部包。

随意修改 module path。module path 一旦被其他代码导入，变化会造成连锁修改。

把所有代码都放在根目录。小项目初期可以简单，但连续演进项目需要尽早建立目录边界。

误以为 `internal/` 只是命名习惯。它是 Go 工具链强制执行的导入限制。

过早引入复杂 workspace。`go work` 解决多模块本地开发问题，不是单模块项目的默认起点。

# 自测问题

1. `go.mod` 中的 module path 会影响哪些导入路径？
2. 为什么 `cmd/gotaskflow/main.go` 不应该直接承载所有业务逻辑？
3. `internal/` 与普通目录相比有什么工具链级别的差异？
4. Day 1 的任务队列为什么不做持久化？
5. 如果未来要增加 HTTP API，应该复用哪一层能力，而不是复制 CLI 逻辑？

# 验收标准

1. `Days/Day01/README.md` 包含完整教程和所有必需章节。
2. `Days/Day01/Day01_module_layout.go` 可以通过 `gofmt`。
3. `go run Days/Day01/Day01_module_layout.go` 可以成功执行。
4. 代码能展示 GoTaskFlow 的模块边界、CLI 入口职责和最小任务队列模型。
5. 本日内容只基于 syllabus Day 1 的主题、重点、实践、理解和参考源展开。

# 明日衔接

Day 2 将进入类型系统、零值与领域建模。今天的 `Job` 和 `Queue` 只是轻量原型；明天会把 `Job`、`Queue`、`Worker`、`JobStatus` 等核心领域模型设计得更严格，重点处理 struct、method、value/pointer receiver、常量枚举、时间建模和领域不变量。

# GitHub Commit

- 仓库：/Users/liushan/Documents/Codex/Golang-Learning
- 代码目录：`Days/Day01/`
- Commit：`待提交`
- Commit 链接：待提交
- Push 状态：待提交
