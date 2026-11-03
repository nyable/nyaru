# nyaru

一个用 Go 编写的 Scoop 交互式 CLI 包装器。查询走 `sfsu` 或 `scoop`，安装、卸载、更新等写操作仍交给 `scoop` 执行。

## 功能特性

- **统一 TUI**：单个 Bubble Tea 程序，列表过滤、多选、详情、确认弹窗、进度和结果提示使用同一套主题和键位。
- **显式后端**：`auto` / `sfsu` / `scoop` 三种模式，行为可预测；sfsu 不可用时自动回退 scoop。
- **批量操作**：安装、卸载、更新、锁定、清理均支持多选；缓存与旧版本列表显示版本、占用大小和可释放空间。
- **Scoop 语法对齐**：`update`、`cleanup`、`config`、`export/import` 等命令尽量保持与 Scoop 一致。
- **自更新**：通过 Scoop 安装时，`nyaru self-update` 会在当前进程退出后执行 `scoop update nyaru`，避免 Windows 文件占用。
- **可视化设置**：`nyaru settings` 或主菜单“设置”可修改后端模式、批量确认和状态提示时长。
- **脚本友好**：查询命令支持 `--json` 和 `--no-tui`；非 TTY 环境自动使用纯文本输出。

## 环境要求

- Windows 10/11
- PowerShell 5.1 或 PowerShell 7+
- [Scoop](https://scoop.sh/)
- 可选：[sfsu](https://github.com/winpax/sfsu)（启用更快的查询路径）

## 安装

### 从 GitHub Releases 下载

从 [Releases](https://github.com/nyable/nyaru/releases) 下载 `nyaru.exe`，放到 PATH 目录中。

### 通过 Scoop 安装

如果你的 bucket 中已包含 nyaru manifest，可以直接安装：

```powershell
scoop install nyaru
```

否则请使用 Releases 中的 `nyaru.exe`，或参考 Scoop 文档添加自定义 bucket。

### 从源码构建

```powershell
go build -o nyaru.exe ./cmd/nyaru
```

## 命令

### 查询与展示

| 命令 | 别名 | 说明 |
|------|------|------|
| `search [query]` | `find`, `query`, `s` | 搜索可安装应用；TUI 中可多选安装 |
| `list [query]` | `ls` | 列出已安装应用 |
| `status` | - | 检查可更新/异常应用 |
| `info <app>` | `detail` | 查看应用详情 |
| `bucket [list]` | `bt`, `ls` | 查看存储桶 |
| `cache [list]` | `ls`, `show` | 查看缓存；按应用聚合 |
| `version` | `v`, `ver` | 显示 nyaru、Scoop、sfsu 版本 |
| `which <command>` | - | 查找 Scoop shim |
| `prefix <app>` | - | 显示应用安装目录 |
| `cat <app>` | - | 显示 manifest |
| `depends <app>` | - | 列出依赖 |
| `home <app>` | - | 打开应用主页 |

### 变更操作

| 命令 | 别名 | 说明 |
|------|------|------|
| `install <app>...` | `add`, `i` | 安装应用；支持 `-g`、`--arch` |
| `uninstall <app>...` | `rm`, `remove` | 卸载应用；支持 `-g`、`--purge` |
| `update [app...]` | - | 无参数更新 Scoop 与 bucket；指定应用更新应用；`*`/`--all` 更新全部本地应用，加 `--global` 同时更新全局应用 |
| `hold <app>...` | `pin` | 锁定应用 |
| `unhold <app>...` | `unpin` | 解锁应用 |
| `cleanup [app...]` | - | 清理旧版本；无参数时在 TUI 中列出可清理项；支持 `--all` |
| `bucket add <name> [source]` | - | 添加存储桶 |
| `bucket rm <name>` | `remove` | 删除存储桶 |
| `cache rm [app...]` | `remove` | 删除缓存；支持 `--all` |
| `reset [app...]` | - | 重置应用；支持 `--all` |
| `config [list]` | - | 查看 Scoop 配置 |
| `config get/set/rm` | - | 读取、设置、删除配置项 |
| `export [path]` | - | 导出 Scoopfile |
| `import <path>` | - | 导入 Scoopfile |
| `checkup` | `doctor` | 检查 Scoop 环境 |
| `scoop <args...>` | - | 原样透传给 scoop，覆盖长尾命令 |

### nyaru 设置

| 命令 | 别名 | 说明 |
|------|------|------|
| `settings` | `prefs`, `preferences` | TTY 下打开设置 TUI；否则列出设置 |
| `settings list` | `ls`, `show` | 列出 nyaru 设置及配置文件路径 |
| `settings get <key>` | - | 读取一个设置项 |
| `settings set <key> <value>` | - | 修改并保存一个设置项 |
| `settings reset` | - | 恢复默认设置 |
| `settings path` | - | 显示 nyaru 配置文件路径 |

注意：`nyaru config` 操作的是 **Scoop** 配置；nyaru 自身的设置使用 `nyaru settings`。

### 入口与自更新

| 命令 | 别名 | 说明 |
|------|------|------|
| `interactive` | `gui`, `tui` | 进入交互式主菜单 |
| 无参数 | - | 等同 `interactive` |
| `self-update` | - | 更新 nyaru 自身；`--check` 只检查 |
| `completion <shell>` | - | 生成 bash/zsh/fish/powershell 补全 |

## 使用示例

```powershell
# 搜索并交互式安装
nyaru search aria2

# 安装多个应用
nyaru install git curl wget --yes

# 查看更新并选择
nyaru status

# 更新 Scoop 自身与 bucket
nyaru update

# 更新全部本地应用
nyaru update --all

# 更新全部本地和全局应用
nyaru update --all --global

# 查看可清理的旧版本并选择清理
nyaru cleanup

# 打开 nyaru 设置界面
nyaru settings

# 脚本方式修改设置
nyaru settings set mode scoop
nyaru settings get status_timeout

# 更新 nyaru 自身
nyaru self-update --check
nyaru self-update

# 脚本读取 JSON
nyaru search git --json --no-tui

# 未封装的 Scoop 命令
nyaru scoop bucket known
```

## TUI 按键

| 按键 | 行为 |
|------|------|
| `↑` / `↓` / `j` / `k` | 移动 |
| `Space` | 选中 / 取消 |
| `a` | 全选当前过滤结果 |
| `A` | 全选全部 |
| `c` | 清空选择 |
| `Enter` | 主操作 / 打开动作菜单 |
| `i` | 详情 |
| `/` | 过滤 |
| `o` | 打开主页 / 源地址 |
| `r` | 刷新 |
| `?` | 帮助 |
| `Esc` / `q` | 返回 |
| `Ctrl+C` | 退出 |

安装、更新、卸载、清理旧版本、删除缓存等操作完成后会进入“执行结果”页，保留运行命令、退出码以及 stdout/stderr 输出；按 `Enter` 或 `Esc` 关闭结果页并返回列表，便于排查失败原因。

主菜单中的“设置”可以修改 backend 模式、批量操作确认和底部状态提示时长，修改后立即写入平台的 nyaru 配置目录。

## 配置

配置文件遵循 XDG 约定，和 Scoop、mise、opencode 等开发工具保持一致：

| 平台 | 路径 |
|------|------|
| 所有平台 | `$XDG_CONFIG_HOME/nyaru/config.yaml`，未设置时使用 `~/.config/nyaru/config.yaml` |
| Windows | 默认即 `%USERPROFILE%\.config\nyaru\config.yaml` |

旧的 `~/.nyaru/config.yaml` 以及 `%APPDATA%\nyaru\config.yaml` 会在首次运行时自动迁移到新位置。

```yaml
mode: auto           # auto | sfsu | scoop
language: zh         # 预留
confirm_batch: true  # 批量操作前确认
status_timeout: 5    # 底部状态提示保留秒数；0 表示不自动清除
```

| 设置项 | 说明 |
|------|------|
| `mode` | 查询后端：`auto` 默认优先 sfsu；也可强制 `sfsu` 或 `scoop` |
| `confirm_batch` | 批量安装、更新、清理前是否显示确认提示 |
| `status_timeout` | 底部状态提示自动消失时间，单位秒；`0` 表示不自动清除 |

`mode` 修改后需要重启 nyaru 才会切换当前会话使用的后端；其他设置在保存后立即生效。也可以直接运行 `nyaru settings` 在 TUI 中修改。

## 自更新说明

- 通过 Scoop 安装时，`self-update` 会启动一个等待当前进程退出的 helper；helper 随后刷新 Scoop 与 buckets，再执行 `scoop update nyaru`。
- 更新日志写入 nyaru 配置目录下的 `self-update.log`（例如 Windows 为 `%USERPROFILE%\.config\nyaru\self-update.log`）。
- 能否更新取决于 Scoop bucket 中是否存在更新的 nyaru manifest；如果 bucket 未发布新版本，会提示已是最新。
- 独立安装（不是通过 Scoop）只做版本检查与下载提示，不自动替换 exe。
- 全局安装使用 `self-update --global`，需要管理员权限。

## 故障排查

### 更新时提示 `Get-FileHash` 不是可识别的 cmdlet

这通常出现在从 PowerShell 7 启动 nyaru 的场景：nyaru 调用 Windows PowerShell 5.1 执行 Scoop 时继承了 PowerShell 7 的 `PSModulePath`，5.1 会优先去 PowerShell 7 的模块目录查找 `Microsoft.PowerShell.Utility`，导致 `Get-FileHash` 无法自动加载。

nyaru 在调用 `powershell.exe` 时已自动移除继承的 `PSModulePath`，让 Windows PowerShell 使用自己的默认模块路径。如果仍然出现，可以先用下面的命令确认环境：

```powershell
powershell.exe -NoProfile -Command "Get-Command Get-FileHash"
```

也可以在启动 nyaru 前临时清理变量：

```powershell
$env:PSModulePath = $null
nyaru
```

## 项目结构

```
nyaru/
├── cmd/nyaru/              # 程序入口
├── internal/
│   ├── cli/                # Cobra 命令与输出
│   ├── config/             # 配置加载与后端模式解析
│   ├── core/               # Runner、后端接口、scoop/sfsu 实现
│   │   └── parse/          # JSON/文本解析器与测试
│   ├── models/             # AppInfo / BucketResult / CacheResult / AppDetail
│   ├── selfupdate/         # 自更新检测与 helper
│   ├── tui/                # 单程序 TUI、主题、列表、详情
│   └── utils/              # 大小转换、浏览器等通用工具
└── .github/workflows/      # CI 与发布
```

## 开发

```powershell
go build ./cmd/nyaru
go test ./...
go vet ./...
```

发布：推送 `v*` 标签会触发 GitHub Actions 构建 Windows AMD64 版本并创建 Release。
