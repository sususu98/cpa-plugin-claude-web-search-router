# Claude Code Web Search 路由插件（CPA）

面向 **Claude Code 内置 `web_search`** 请求的独立 CPA 插件。仓库结构与发布流程参考 [router-for-me/cpa-plugin-gemini-cli](https://github.com/router-for-me/cpa-plugin-gemini-cli)。

- 仓库：<https://github.com/sususu98/cpa-plugin-claude-web-search-router>
- 插件 ID：`claude-web-search-router`（`plugins.configs` 的 key 必须与之一致）

## 做什么

Claude Code 会发带 `web_search_20250305` / `web_search_20260209` 的 Claude Messages 请求。本插件通过 **ModelRouter** 识别这类流量，再由插件 **Executor** 把搜索交给其它后端执行，并把结果包装成 Claude 能识别的形态（`server_tool_use`、`web_search_tool_result` 等）。

典型能力：

- 同时注册 **ModelRouter + Executor**
- 识别协议 `claude` / `anthropic` 下的 Claude Code websearch 特征
- 默认 **fallback** 链：`antigravity → codex → xai → tavily`
- **`route` 写成 YAML 列表** 时，按列表顺序 fallback（支持别名 `antigravity` / `codex` / `xai`；未知项会跳过并去重）
- 单请求内对 **429 / 503 / 502** 自动换下一个后端；失败次数多的后端会在后续请求中被降权（内存 penalty，无需配置）
- **Tavily** 路径在插件内直接搜网，并合成 Claude SSE / JSON 响应

## 依赖

- CPA 宿主 **v7.2.31+**（`github.com/router-for-me/CLIProxyAPI/v7`）
- 构建需 **CGO**（`c-shared` 动态库）

## 构建

```bash
go test ./...
mkdir -p dist
CGO_ENABLED=1 go build -buildmode=c-shared \
  -o dist/claude-web-search-router.dylib \
  ./cmd/claude-web-search-router
```

- macOS：`.dylib`
- Linux：`.so`（输出文件名自行改为 `claude-web-search-router.so`）
- Windows：`.dll`

打 tag `v*` 推送后，GitHub Actions 会按多平台打包（见 `.github/workflows/build.yml`）。

## CPA 配置示例

```yaml
plugins:
  path:
    - /绝对路径/dist/claude-web-search-router.dylib
  configs:
    claude-web-search-router:
      enabled: true
      priority: 20
      route: fallback
      antigravity_model: "" # 空：从 registry 找 supports_web_search
      codex_model: "gpt-5.4-mini"
      xai_model: "grok-4.3"
      tavily_api_keys:
        - "tvly-xxxxxxxx"
      require_web_search_only: true
```

不写 `route` 时默认为 `fallback`。

### 自定义 fallback 顺序

```yaml
plugins:
  configs:
    claude-web-search-router:
      enabled: true
      route:
        - tavily
        - codex_web_search
        - xai_web_search
      tavily_api_keys:
        - "tvly-xxxxxxxx"
```

`route` 为标量 `fallback`、省略、或空时，仍使用默认顺序 `antigravity → codex → xai → tavily`。

### 最小配置（主要靠 Tavily 兜底）

```yaml
plugins:
  configs:
    claude-web-search-router:
      enabled: true
      route: fallback
      tavily_api_keys:
        - "tvly-xxxxxxxx"
      require_web_search_only: true
```

### 固定单一后端

| `route`              | 说明                                                                                                   |
| -------------------- | ------------------------------------------------------------------------------------------------------ |
| `antigravity_google` | 走 Antigravity 原生 googleSearch（需宿主有 antigravity 且模型可解析）                                  |
| `codex_web_search`   | 走 Codex Responses `web_search`（默认模型 `gpt-5.4-mini`，**不会**把客户端 Claude 模型名转发给 Codex） |
| `xai_web_search`     | 走 xAI Responses `web_search`（默认 `grok-4.3`）                                                       |
| `tavily`             | 仅插件内 Tavily                                                                                        |
| `default_provider`   | 走 CPA 内置 `default_provider` / `default_provider_model`（无 fallback 编排）                          |

## 识别规则

满足以下条件才会接管（`enabled: true` 时）：

1. `SourceFormat` 为 `claude` 或 `anthropic`
2. `tools[]` 含 `web_search_20250305` 或 `web_search_20260209`
3. 若 `require_web_search_only: true`（默认），则 tools 中只能有上述 websearch 类型
4. 另满足其一即可：
   - system / user 文案像 Claude Code（如 “web search tool use”、用户以 `Perform a web search for the query:` 开头）
   - 或 tools 仅有 typed websearch（见上）

搜索词提取顺序：优先从 `Perform a web search for the query:` 解析，否则取最近一条 user 文本。

## 配置字段

| 字段                                          | 说明                                           |
| --------------------------------------------- | ---------------------------------------------- |
| `enabled`                                     | `false` 时对所有匹配请求返回不处理             |
| `priority`                                    | ModelRouter 优先级（越大越先匹配）             |
| `route`                                       | 见上表；默认 `fallback`                        |
| `antigravity_model`                           | Antigravity 执行模型；勿填客户端 Claude 模型名 |
| `codex_model`                                 | Codex 模型；空 → `gpt-5.4-mini`                |
| `xai_model`                                   | xAI 模型；空 → `grok-4.3`                      |
| `default_provider` / `default_provider_model` | 仅 `route=default_provider`                    |
| `tavily_api_keys`                             | `tavily` 或 fallback 最后一步必填；多 key 轮询 |
| `require_web_search_only`                     | `true` 更贴近 Claude Code 独占 websearch       |

## xAI / Codex 说明（与上游一致）

- xAI 服务端 `web_search` 文档模型为 `grok-4.3`；插件不会在 `xai_model` 为空时把 `claude-sonnet-4-6` 发给 xAI。
- Claude 的 `allowed_domains` 经 CPA translator 可映射到 Responses `filters.allowed_domains`；`blocked_domains` → `excluded_domains` 目前**未**完整映射（以 CPA 主仓 translator 为准）。

## CI 与发版

- 推送到 `main` 或开 PR：只跑 `go test` / `go vet`
- **推送 tag `v*`（例如 `v0.1.0`）**：触发全平台构建，并自动创建 GitHub Release（含各平台 zip 与 `checksums.txt`）
- 矩阵含 **linux / darwin / windows（amd64+arm64）** 与 **freebsd amd64**（与 [cpa-plugin-gemini-cli](https://github.com/router-for-me/cpa-plugin-gemini-cli) 一致）
- 也可在 Actions 页手动 **Run workflow**

```bash
git tag v0.1.0
git push origin v0.1.0
```

tag 必须以 **`v` 开头**（与 workflow 中 `tags: v*` 一致）。

## 开发与测试

```bash
go test ./...
go vet ./...
```

可选：将 Claude Code 抓包 JSON 放到 `testdata/claude_code_web_search_capture.json`，`detect_test` 会用该 fixture 做识别测试；无文件时相关测试会 skip。

## 许可证

MIT，见 [LICENSE](LICENSE)。
