package smtp

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	"sync"
)

// Rendering
type RenderFunc func(subject string, data any) ([]byte, error)

func Render(fs fs.FS, filenames ...string) (render RenderFunc, err error) {
	var (
		init sync.Once

		tpl *template.Template
		sb  *strings.Builder
	)

	init.Do(func() { tpl, err = template.ParseFS(fs, filenames...) })

	render = func(subject string, data any) ([]byte, error) {
		sb = &strings.Builder{}
		if err = tpl.Execute(sb, data); err != nil {
			return nil, err
		}

		s := fmt.Sprintf("Subject: %s\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n%s", subject, sb.String())

		return []byte(s), nil
	}

	return
}
