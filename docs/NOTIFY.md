# 订单推送说明（飞书 / 微信）

下单成功后可自动将订单推送到已配置的渠道；也可手动对历史订单再次推送。

## 支持的渠道

| 渠道 | 环境变量 | 说明 |
| --- | --- | --- |
| **飞书** | `FEISHU_WEBHOOK_URL` | 飞书群 → 设置 → 群机器人 → 自定义机器人 → Webhook 地址 |
| **企业微信** | `WECOM_WEBHOOK_URL` | 企业微信群机器人 Webhook（商户/后厨常用） |
| **个人微信** | `PUSHPLUS_TOKEN` | 通过 [PushPlus](https://www.pushplus.plus/) 推送到个人微信，注册后获取 token |

可同时配置多个渠道，新订单会依次推送。

> 个人微信没有官方简单的 Webhook，推荐使用 **PushPlus** 或 **企业微信群机器人**。

## HTTP 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/notify/status` | 查看已配置的推送渠道 |
| POST | `/api/notify/test` | 发送测试消息 |
| POST | `/api/orders/:id/notify` | 手动推送指定订单 |

下单 `POST /api/orders` 成功后，若已配置渠道，会**异步自动推送**。

## 配置示例

```bash
# 飞书
export FEISHU_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/xxxx"

# 企业微信
export WECOM_WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxx"

# 个人微信（PushPlus）
export PUSHPLUS_TOKEN="your_pushplus_token"

# 小票/推送抬头（可选）
export SHOP_NAME="小馆点餐"
```

Docker Compose 可在项目根目录 `.env` 中配置上述变量。

## 推送内容示例

```
【小馆点餐】新订单 #3

单号: #3
时间: 2026-05-31 21:30:00
状态: 待处理
—— 明细 ——
· 西红柿炒鸡蛋 × 2
· 米饭 × 1
合计: 3 件
备注: 不要香菜
```

## 测试

```bash
curl http://localhost:8080/api/notify/status
curl -X POST http://localhost:8080/api/notify/test
curl -X POST http://localhost:8080/api/orders/1/notify
```
