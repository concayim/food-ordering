# 小馆点餐系统 (Golang + Vue3)

一个简单的点餐程序，支持菜品上下架、原材料管理（库存可设为无限）、下单与订单管理。

## 功能

- **点餐**：浏览已上架菜品、**点击菜品查看烹饪方式与原材料详情**、加入购物车、下单（下单时按菜品配方自动扣减原材料库存）。已去除价格，下单只统计件数
- **随机轮盘**：从已上架菜品中转盘随机抽取一道（"今天吃什么"）。带**转动音效**（每过一个扇区"嗒"声 + 抽中后小段旋律，可静音）、**转动时高亮当前扇区**，**抽中后自动加入购物车并跳转到点餐页**；购物车跨页签共享
- **菜品 / 汤品管理**：录入 / 编辑 / 删除，一键上架 / 下架；支持
  - **类型**：菜品 / 汤品
  - **拍照录入**：调用摄像头拍照或上传图片作为菜品图（图片存于后端 `uploads/`）
  - **原材料手动关联**：从已有原材料中选择，或直接**手动新建原材料**并关联（用量自填）
  - **烹饪方法**（上架前可填写）：可自己录入；不懂做法时可点「✨ AI 推荐」由大模型生成，或粘贴**来源网站菜谱网址**点「🌐 从网址获取」自动抓取并提炼（见下方 AI 配置）
- **原材料 / 库存**：管理原材料，库存可设为具体数量或 **无限（∞）**；无限库存下单时不扣减
- **订单记录**：查看历史订单，更新订单状态（待处理 / 已支付 / 完成 / 取消）
- **项目文档**：系统内可视化展示需求文档、更新日志、REST 接口文档及在线调试

## 技术栈

- 后端：Go + Gin + GORM + SQLite（纯 Go 驱动 glebarez/sqlite，无需 CGO）
- 前端：Vue 3 + Vite

## 运行

### 1. 启动后端（端口 8080）

```bash
cd backend
go run .
```

首次运行会自动创建 `food.db` 并写入演示数据（含一道汤品「紫菜蛋花汤」）。

#### 配置 AI 烹饪方法推荐（可选）

「✨ AI 推荐」与「🌐 从网址获取」会调用 OpenAI 风格的 `chat/completions` 接口。**密钥请勿写入代码库**，通过环境变量配置：

```bash
cp backend/.env.example backend/.env
# 编辑 backend/.env 填入 AI_API_KEY、AI_BASE_URL、AI_MODEL
export $(grep -v '^#' backend/.env | xargs)
cd backend && go run .
```

| 变量 | 说明 |
| --- | --- |
| `AI_API_KEY` | 必填（启用 AI 时），你的模型 appkey |
| `AI_BASE_URL` | 选填，默认 `https://api.openai.com/v1` |
| `AI_MODEL` | 选填，默认 `gpt-4o-mini` |

`/ai/cooking-method-from-url` 会抓取传入网址的网页正文（最多 1MB），再交给模型提炼烹饪步骤。

### 2. 启动前端（端口 5173）

```bash
cd frontend
npm install
npm run dev
```

浏览器打开 http://localhost:5173 （前端已配置 `/api` 代理到后端 8080）。

## 文档

| 文件 | 说明 |
| --- | --- |
| [CHANGELOG.md](./CHANGELOG.md) | 版本更新日志 |
| [docs/REQUIREMENTS.md](./docs/REQUIREMENTS.md) | 需求文档 |
| 系统内「项目文档」页 | 需求 /  changelog / 接口可视化与在线调试 |

## 主要接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/dishes?onShelf=true` | 菜品列表（可只看上架） |
| POST | `/api/dishes` | 新增菜品 |
| PUT | `/api/dishes/:id` | 编辑菜品 |
| PATCH | `/api/dishes/:id/shelf` | 切换上架/下架 |
| DELETE | `/api/dishes/:id` | 删除菜品 |
| GET/POST | `/api/ingredients` | 原材料列表 / 新增 |
| PATCH | `/api/ingredients/:id/stock` | 设置库存（`{"infinite":true}` 设为无限） |
| GET/POST | `/api/orders` | 订单列表 / 下单 |
| PATCH | `/api/orders/:id/status` | 更新订单状态 |
| POST | `/api/upload` | 上传图片（multipart，字段名 `file`），返回 `{url}` |
| POST | `/api/ai/cooking-method` | AI 生成烹饪方法（需配置 `AI_API_KEY`） |
| POST | `/api/ai/cooking-method-from-url` | 从来源网站抓取并提炼烹饪方法 |

> 库存约定：`stock = -1` 表示无限库存。
