# 个人版维护与更新

个人仓库：https://github.com/2tdbbfgk2m-beep/cursor-byok

上游仓库：https://github.com/Sxuan-Coder/cursor-byok

## 保留的修改

- 提供商可统一命名；所属模型自动显示「提供商名称 / 模型名称」。
- 提供商改名不改变渠道 ID 或上游模型标识，新增模型继承提供商名称。
- 本机助手从个人仓库获取更新；模型地址、密钥和个人配置不进入仓库。

## 自动发布

推送 main 或手动运行 Personal macOS Release 后，CI 使用 macOS 15 Intel runner 交叉构建 Apple Silicon 应用（产物通过 lipo 验证为 arm64）、运行回归测试、临时签名并验证、生成 SHA-256 更新清单。全部完成后才发布为 latest。

版本格式：`上游版本-personal.流水线编号`，例如 `1.0.1-personal.10`。编号按数值比较，10 大于 9。`build/config.yml` 保持上游基础版本，个人后缀由 CI 注入，不把构建产物提交回 main。

更新源：`https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/latest/download/update.json`。

产物：`cursor-byok-版本-macos-arm64.tar.gz` 和 `update.json`。应用 ID 与安装位置延续原版，因此使用现有配置目录。构建是 ad-hoc 签名，没有 Apple 公证。

上游的多平台 Release 工作流限定仅在上游仓库运行，此 fork 使用 Personal macOS Release 和 Sync Upstream；发布后也在 GitHub Actions 中禁用继承工作流。

## 自动合并上游

Sync Upstream 每天 UTC 03:17（北京时间 11:17）检查上游，也可以在 Actions 手动运行。GitHub 的计划任务可能延迟执行。

1. 上游没有新提交：结束，不发布重复版本。
2. 存在新提交：先在临时 runner 上合并，保留个人提交。
3. 有冲突：中止合并，工作流失败，不修改远端 main 和已发布版本；在 Actions 查看失败日志，人工解决冲突后推送 main。
4. 无冲突：回归测试通过后推送 main，再显式触发个人发布流程。GITHUB_TOKEN 触发的普通 push 不会自动触发另一条工作流，因此这里使用 workflow_dispatch。
5. 发布构建失败：保留已有 latest 版本；修复 main 后重跑发布。不会强推或覆盖已有同名 Release。

首次启用后，可手动运行 `gh workflow run sync-upstream.yml --repo 2tdbbfgk2m-beep/cursor-byok --ref main` 验证没有上游变更时安全结束。

## 手工维护

在有完整历史的本地 checkout 中：

```sh
git fetch origin main
git switch main
git pull --ff-only origin main
git fetch upstream main
git merge upstream/main
# 解决冲突后运行相关回归测试，再推送个人 main。
git push origin main
```

不要把上游原版二进制直接覆盖到此个人版，否则会丢失个人更新源和功能。需要官方新功能时，先把源码合入个人 main，由 CI 重新构建。

## 恢复

更新前先备份 `~/.cursor-local-assistant-v2/config.yaml`。要回退个人版本，从本仓库 Releases 下载上一个已发布的 arm64 包，退出助手后替换 `/Applications/Cursor助手.app`。保留配置目录；重启后核对 provider 名称和模型选择。无需清除账号或密钥。
