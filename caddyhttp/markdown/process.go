package markdown

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/mholt/caddy/caddyhttp/markdown/metadata"
	"github.com/mholt/caddy/caddyhttp/markdown/summary"
	"github.com/russross/blackfriday"
)

type FileInfo struct {
	os.FileInfo
	ctx httpserver.Context
}

func (f FileInfo) Summarize(wordcount int) (string, error) {
	fp, err := f.ctx.Root.Open(f.Name())
	if err != nil {
		return "", err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", err
	}

	return string(summary.Markdown(buf, wordcount)), nil
}

// Markdown processes the contents of a page in b. It parses the metadata
// (if any) and uses the template (if found).
func (c *Config) Markdown(title string, r io.Reader, dirents []os.FileInfo, ctx httpserver.Context) ([]byte, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	parser := metadata.GetParser(body)
	markdown := parser.Markdown()
	mdata := parser.Metadata()

	// process markdown
	extns := 0
	extns |= blackfriday.EXTENSION_TABLES
	extns |= blackfriday.EXTENSION_FENCED_CODE
	extns |= blackfriday.EXTENSION_STRIKETHROUGH
	extns |= blackfriday.EXTENSION_DEFINITION_LISTS
	html := blackfriday.Markdown(markdown, c.Renderer, extns)

	// set it as body for template
	mdata.Variables["body"] = string(html)

	// fixup title
	mdata.Variables["title"] = mdata.Title
	if mdata.Variables["title"] == "" {
		mdata.Variables["title"] = title
	}

	// massage possible files
	files := []FileInfo{}
	for _, ent := range dirents {
		file := FileInfo{
			FileInfo: ent,
			ctx:      ctx,
		}
		files = append(files, file)
	}

	return execTemplate(c, mdata, files, ctx)
}