package infra

import (
	"fmt"
	"os"
	"strings"
)

type ReportGenerator struct {
	Markdown *string
}

func NewReportGenerator() ReportGenerator {
	s := ""

	return ReportGenerator{
		Markdown: &s,
	}
}

func (g ReportGenerator) Append(msg string) {
	*g.Markdown += fmt.Sprintf("%v\n", msg)
}

func (g ReportGenerator) BulletList(items []string, depth int) {
	for _, item := range items {
		*g.Markdown += fmt.Sprintf("%v- %v\n", strings.Repeat(" ", depth*4), item)
	}
}

func (g ReportGenerator) WriteFile(path string) error {
	if err := os.WriteFile(path, []byte(*g.Markdown), 0644); err != nil {
		return err
	}
	*g.Markdown = ""

	return nil
}
