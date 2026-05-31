# 打印机接口说明

系统预留打印机驱动接口，默认使用 **stub** 模拟打印（返回小票文本、写入日志）。接入实体打印机时，只需实现 `PrinterDriver` 接口并注册驱动名即可。

## HTTP 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/printer/status` | 查询打印机驱动状态 |
| POST | `/api/printer/test` | 测试打印（联调用） |
| GET | `/api/orders/:id/receipt` | 预览小票（JSON + 文本，不打印） |
| POST | `/api/orders/:id/print` | 提交订单到打印机 |

### 示例

```bash
# 状态
curl http://localhost:8080/api/printer/status

# 打印订单 #1
curl -X POST http://localhost:8080/api/orders/1/print
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PRINTER_DRIVER` | `stub` | 驱动名称，后续可扩展 `escpos` 等 |
| `SHOP_NAME` | `小馆点餐` | 小票抬头店名 |

## 驱动扩展

在 `backend/printer.go` 中实现接口：

```go
type PrinterDriver interface {
    Name() string
    Status() PrinterStatus
    Print(receipt Receipt) PrintResult
}
```

在 `initPrinter()` 中按 `PRINTER_DRIVER` 注册具体实现，例如：

- ESC/POS 热敏打印机（USB / 串口 / 网络）
- 系统 CUPS 打印
- 云打印服务

小票文本格式由 `formatReceiptText()` 生成，驱动可直接发送该文本或自行排版。
