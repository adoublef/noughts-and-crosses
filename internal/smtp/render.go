package smtp

import (
	"bytes"
	"html/template"
	"io/fs"
	"sync"
)

// RenderFunc returns a mail type
type RenderFunc func(data any, subj string, to ...string) (*Mail, error)

func Render(fs fs.FS, filenames ...string) (render RenderFunc, err error) {
	var (
		init sync.Once

		tpl *template.Template
		sb  *bytes.Buffer
	)

	init.Do(func() { tpl, err = template.ParseFS(fs, filenames...) })

	render = func(data any, subj string, to ...string) (*Mail, error) {
		sb = &bytes.Buffer{}
		if err = tpl.Execute(sb, data); err != nil {
			return nil, err
		}

		m := &Mail{
			To:   to,
			Subj: subj,
			Body: sb.Bytes(),
		}

		return m, nil
	}

	return
}
