package infra

import (
	"fmt"
	"os"
)

type ReportGenerator struct {
	Markdown string
}

func (g *ReportGenerator) Append(msg string) {
	g.Markdown += fmt.Sprintf("%v\n", msg)
}

func (g *ReportGenerator) WriteFile(path string) error {
	if err := os.WriteFile(path, []byte(g.Markdown), 0644); err != nil {
		return err
	}
	g.Markdown = ""

	return nil
}
