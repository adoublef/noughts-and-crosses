package smtp

import (
	"bytes"
	"html/template"
	"io/fs"
	"sync"
)

// Rendering
type RenderFunc func(m *Mail, data any) error

func Render(fs fs.FS, filenames ...string) (render RenderFunc, err error) {
	var (
		init sync.Once

		tpl *template.Template
		sb  *bytes.Buffer
	)

	init.Do(func() { tpl, err = template.ParseFS(fs, filenames...) })

	render = func(m *Mail, data any) error {
		sb = &bytes.Buffer{}
		if err = tpl.Execute(sb, data); err != nil {
			return err
		}

		m.Body = sb.Bytes()
		return nil
	}

	return
}
