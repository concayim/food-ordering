package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// PrinterDriver 打印机驱动接口，后续可接入 ESC/POS、串口、网络打印机等实现。
type PrinterDriver interface {
	Name() string
	Status() PrinterStatus
	Print(receipt Receipt) PrintResult
}

// PrinterStatus 打印机状态
type PrinterStatus struct {
	Ready   bool   `json:"ready"`
	Driver  string `json:"driver"`
	Message string `json:"message"`
}

// ReceiptLine 小票一行明细
type ReceiptLine struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// Receipt 打单内容
type Receipt struct {
	OrderID   uint          `json:"orderId"`
	ShopName  string        `json:"shopName"`
	Status    string        `json:"status"`
	Remark    string        `json:"remark"`
	Items     []ReceiptLine `json:"items"`
	CreatedAt time.Time     `json:"createdAt"`
}

// PrintResult 打印结果
type PrintResult struct {
	Success bool   `json:"success"`
	JobID   string `json:"jobId"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"` // stub 模式下返回模拟小票文本
}

var printer PrinterDriver

func initPrinter() {
	switch strings.ToLower(os.Getenv("PRINTER_DRIVER")) {
	case "stub", "":
		printer = &stubPrinter{}
	default:
		// 未知驱动名时仍用 stub，并在状态中提示
		printer = &stubPrinter{unknownDriver: os.Getenv("PRINTER_DRIVER")}
	}
	log.Printf("打印机驱动: %s (%s)", printer.Name(), printer.Status().Message)
}

// ---------- stub 驱动（默认）----------

type stubPrinter struct {
	unknownDriver string
}

func (p *stubPrinter) Name() string {
	if p.unknownDriver != "" {
		return "stub"
	}
	return "stub"
}

func (p *stubPrinter) Status() PrinterStatus {
	if p.unknownDriver != "" {
		return PrinterStatus{
			Ready:   false,
			Driver:  "stub",
			Message: fmt.Sprintf("驱动 %q 尚未实现，当前为 stub 模拟打印", p.unknownDriver),
		}
	}
	return PrinterStatus{
		Ready:   true,
		Driver:  "stub",
		Message: "模拟打印模式，小票内容以文本返回，接入真实驱动后可直接出纸",
	}
}

func (p *stubPrinter) Print(receipt Receipt) PrintResult {
	content := formatReceiptText(receipt)
	jobID := fmt.Sprintf("stub-%d-%d", receipt.OrderID, time.Now().UnixNano())
	log.Printf("[printer] 订单 #%d 模拟打印:\n%s", receipt.OrderID, content)
	return PrintResult{
		Success: true,
		JobID:   jobID,
		Message: "已提交模拟打印（stub），请接入真实驱动后自动出纸",
		Content: content,
	}
}

func shopName() string {
	if n := os.Getenv("SHOP_NAME"); n != "" {
		return n
	}
	return "小馆点餐"
}

func formatReceiptText(r Receipt) string {
	var b strings.Builder
	line := strings.Repeat("-", 32)
	b.WriteString(centerText(r.ShopName, 32) + "\n")
	b.WriteString(centerText("订单小票", 32) + "\n")
	b.WriteString(line + "\n")
	b.WriteString(fmt.Sprintf("单号: #%d\n", r.OrderID))
	b.WriteString(fmt.Sprintf("时间: %s\n", r.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("状态: %s\n", orderStatusLabel(r.Status)))
	b.WriteString(line + "\n")
	total := 0
	for _, it := range r.Items {
		b.WriteString(fmt.Sprintf("%s  x%d\n", it.Name, it.Quantity))
		total += it.Quantity
	}
	b.WriteString(line + "\n")
	b.WriteString(fmt.Sprintf("合计: %d 件\n", total))
	if strings.TrimSpace(r.Remark) != "" {
		b.WriteString(fmt.Sprintf("备注: %s\n", r.Remark))
	}
	b.WriteString(line + "\n")
	b.WriteString(centerText("谢谢惠顾", 32) + "\n")
	return b.String()
}

func centerText(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	pad := (width - len(runes)) / 2
	return strings.Repeat(" ", pad) + s
}

func orderStatusLabel(status string) string {
	switch status {
	case "pending":
		return "待处理"
	case "paid":
		return "已支付"
	case "done":
		return "已完成"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}

func loadReceipt(orderID uint) (Receipt, error) {
	var order Order
	if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
		return Receipt{}, err
	}
	items := make([]ReceiptLine, 0, len(order.Items))
	for _, it := range order.Items {
		items = append(items, ReceiptLine{Name: it.DishName, Quantity: it.Quantity})
	}
	return Receipt{
		OrderID:   order.ID,
		ShopName:  shopName(),
		Status:    order.Status,
		Remark:    order.Remark,
		Items:     items,
		CreatedAt: order.CreatedAt,
	}, nil
}
