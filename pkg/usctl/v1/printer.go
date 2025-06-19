package v1

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// TablePrinter 用于格式化输出表格数据
type TablePrinter struct {
	writer *tabwriter.Writer
}

// NewTablePrinter 创建新的表格打印器
func NewTablePrinter(out io.Writer) *TablePrinter {
	return &TablePrinter{
		writer: tabwriter.NewWriter(out, 0, 8, 2, ' ', 0),
	}
}

// PrintHeader 打印表头
func (p *TablePrinter) PrintHeader(columns ...string) {
	// 打印表头
	fmt.Fprintln(p.writer, formatRow(columns...))
}

// PrintRow 打印数据行
func (p *TablePrinter) PrintRow(columns ...string) {
	fmt.Fprintln(p.writer, formatRow(columns...))
}

// Flush 刷新输出
func (p *TablePrinter) Flush() {
	p.writer.Flush()
}

// formatRow 格式化行数据
func formatRow(columns ...string) string {
	var row string
	for i, col := range columns {
		if i == len(columns)-1 {
			row += col
		} else {
			row += col + "\t"
		}
	}
	return row
}
