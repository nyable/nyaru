# nyaru

一个用 Go 语言编写的 CLI 工具，包装 [Scoop](https://scoop.sh/) 包管理器命令，提供 TUI 交互界面，支持多选、过滤、批量操作。

## 功能特性

- **双后端支持**：支持 `sfsu` 和 `scoop` 两种后端，`auto` 模式自动检测
- **TUI 交互界面**：基于 [bubbletea](https://github.com/charmbracelet/bubbletea) 的终端 UI，支持列表浏览、模糊过滤、多选操作
- **批量操作**：支持同时安装/卸载/更新多个应用
- **交互模式**：运行 `nyaru` 或 `nyaru interactive` 进入主菜单

## 安装

### 从 GitHub Releases 下载

从 [Releases](https://github.com/nyable/nyaru/releases) 页面下载 `nyaru.exe`，放置到 PATH 目录中。

### 从源码构建

```bash
go build -o nyaru.exe ./cmd/nyaru
```

## 命令列表

| 命令 | 别名 | 说明 |
|------|------|------|
| `search <query>` | `find`, `query`, `s` | 搜索可安装的应用程序 |
| `install <app>...` | `add`, `i` | 安装一个或多个应用 |
| `uninstall <app>...` | `rm`, `remove` | 卸载一个或多个应用 |
| `list` | `ls` | 列出所有已安装的应用 |
| `status` | - | 检查已安装应用的更新状态 |
| `update [app...]` | - | 更新应用（不指定则更新全部） |
| `info <app>` | `detail` | 查看应用详细信息 |
| `hold <app>...` | `pin` | 锁定应用，阻止更新 |
| `unhold <app>...` | `unpin` | 解锁应用，允许更新 |
| `bucket list` | `ls` | 列出已添加的存储桶 |
| `bucket add <name> [source]` | - | 添加存储桶 |
| `bucket rm <name>` | `remove` | 删除存储桶 |
| `cache list` | `ls`, `show` | 显示缓存内容 |
| `version` | `v`, `ver` | 显示版本信息 |
| `interactive` | `gui`, `tui` | 进入交互模式主菜单 |

## 使用示例

```bash
# 搜索应用
nyaru search aria2

# 安装多个应用
nyaru install git curl wget

# 查看可更新的应用（交互式选择更新）
nyaru status

# 进入交互模式
nyaru
```

## 配置

配置文件位于 `~/.nyaru/config.yaml`：

```yaml
mode: auto  # auto | sfsu | scoop
```

| 模式 | 说明 |
|------|------|
| `auto` | 默认值，自动检测 `sfsu` 是否可用，不可用则回退到 `scoop` |
| `sfsu` | 强制使用 sfsu 后端 |
| `scoop` | 强制使用 scoop 原生命令 |

## 项目结构

```
nyaru/
├── cmd/nyaru/          # 程序入口
├── internal/
│   ├── cli/            # Cobra 命令定义
│   ├── config/         # 配置管理
│   ├── core/           # 包管理器接口实现
│   ├── models/         # 数据模型
│   ├── tui/            # TUI 组件（bubbletea）
│   └── utils/          # 工具函数
└── .github/workflows/  # CI/CD
```

## 开发

### 依赖

- Go 1.24.2+
- [cobra](https://github.com/spf13/cobra) - CLI 框架
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI 组件
- [lipgloss](https://github.com/charmbracelet/lipgloss) - TUI 样式

### 构建

```bash
go build ./cmd/nyaru
```

### 发布

推送 `v*` 标签触发 GitHub Actions 自动构建并创建 Release：

```bash
git tag v0.0.7
git push origin v0.0.7
```