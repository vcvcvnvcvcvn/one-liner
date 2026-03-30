# ol - One-liner Command Assistant

一个自然语言转终端命令的工具，成为用户和 terminal 之间的中间层。

🌐 **[English](README.md) | 简体中文**

---

## 快速安装

### 使用安装脚本（推荐）

```bash
curl -sSL https://raw.githubusercontent.com/vcvcvnvcvcvn/one-liner/main/install.sh | bash
```

### 从源码安装

```bash
# 编译并安装
go build -o ol .
sudo mv ol /usr/local/bin/
```

### 使用 Makefile

```bash
# 编译当前平台
make build

# 安装到系统
make install

# 编译所有平台
make build-all
```

## 跨平台编译

本项目支持多平台交叉编译：

| 平台 | 架构 | 输出文件名 |
|------|------|-----------|
| Linux | amd64 | `ol-linux-amd64` |
| Linux | arm64 | `ol-linux-arm64` |
| macOS | amd64 (Intel) | `ol-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | `ol-darwin-arm64` |
| Windows | amd64 | `ol-windows-amd64.exe` |
| Windows | arm64 | `ol-windows-arm64.exe` |

### 手动交叉编译

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ol-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ol-darwin-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o ol-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o ol-windows.exe .
```

## 使用方法

```bash
# 基本用法
ol <你的自然语言描述>

# 示例
ol 列出最近24小时修改过的所有文件
ol 查找当前目录下大于100MB的文件
ol 显示带图形化可视化的git日志

# 其他命令
ol --help      # 显示帮助
ol --init      # 重新初始化配置
ol --version   # 显示版本
```

## 交互控制

生成命令后，你可以：

| 按键 | 动作 |
|------|------|
| `Enter` | 执行生成的命令 |
| `r` | 重新生成一个不同的命令 |
| `Ctrl+C` | 取消 |

重新生成时，系统会记住之前生成的命令，并要求 AI 提供不同的实现方式。

## 首次使用

首次运行 `ol` 时会自动进入初始化流程，需要选择：

1. **OpenAI Compatible API** - 支持 OpenAI 或兼容的 API（如 Azure、本地模型等）
2. **Anthropic Compatible API** - 支持 Claude API

配置信息会保存在 `~/.ol_config.json`。

## 配置文件示例

### OpenAI 配置
```json
{
  "api_type": "openai",
  "api_key": "sk-...",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4o-mini",
  "thinking": false
}
```

### Anthropic 配置
```json
{
  "api_type": "anthropic",
  "api_key": "sk-ant-...",
  "base_url": "https://api.anthropic.com/v1/messages",
  "model": "claude-3-5-haiku-20241022",
  "thinking": false
}
```

## 特性

- 🤖 支持 OpenAI 和 Anthropic API
- ⏳ 等待响应时显示优雅的转圈圈动画
- 🛡️ 执行前显示命令并等待用户确认
- 🔄 按 `r` 键重新生成不同风格的命令
- 🐚 自动检测并使用当前 shell 执行命令
- 🔒 配置文件权限设置为 0600（仅用户可读）
- 🌐 跨平台支持：Linux、macOS、Windows

## 系统提示词 (System Prompt)

工具会自动检测用户的操作系统和架构，并注入到 system prompt 中，帮助 AI 生成更适合当前平台的命令。

---

开发环境：Go 1.22+
