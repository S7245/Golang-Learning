# Golang-Learning

Golang 50 天高级学习计划仓库。

本仓库的唯一课程事实源是 [Course/syllabus.json](/Users/liushan/Documents/Codex/Golang-Learning/Course/syllabus.json)。每日自动化必须先读取该文件，再生成当天 Notion 教程和 `Days/DayNN/` 代码目录。

## 课程项目

连续项目：`GoTaskFlow`

目标：从本地 CLI 任务队列逐步演进为生产级 HTTP 任务工作流平台，覆盖 Go 语言机制、工程实践、并发、持久化、测试、性能、安全、可观测性、CI 与架构治理。

## 自动化约定

- 开始日期：`2026-05-16`
- 时区：`Asia/Shanghai`
- 每日运行：`12:00`
- dayNumber 计算：`当前日期 - startDate + 1`
- 当 dayNumber 不在 `1..50` 时停止
- 每次运行必须读取 `Course/syllabus.json`
- 当天缺少 `topic` / `focus` / `practice` / `understanding` / `reference` 时停止并报告错误
- 不覆盖旧文件
- 每天创建 `Days/DayNN/`
- 提交信息格式：`Day NN: add {topic} learning code`
- push 失败时，Notion 仍写入课程，并在底部记录失败原因

## 目录约定

```text
.gitignore
README.md
Course/
  syllabus.json
Days/
  Day01/
    README.md
    Day01_<codeSlug>.go
```
