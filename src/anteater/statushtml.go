package anteater

import (
	"html/template"
	"math"
	"fmt"
	"io"
	"time"
)


const (
	tmplMain = `
<html>
	<head>
		<title>{{.Title}}</title>
	</head>
	<body>
	{{with .Tables}}
		{{range .}}
		<div>
			<h3>{{.Title}}</h3>
			<table>
			{{with .Values}}
			    {{range .}}
				<tr>
					<td style="text-align:right">{{.Key}}:</td>
					<td><b>{{.Value}}</b></td>
				</tr>	
			    {{end}}
			{{end}}
			</table>
			</div>
		{{end}}
	{{end}}
	</body>
</html>`
	
	tmplTable = 
`		
`
)

type KeyValue struct {
	Key, Value string
}

type HtmlMain struct {
	Title string
	Tables []*HtmlTable
}

type HtmlTable struct {
	Title string
	Values []*KeyValue
}

var (
	TmplMain *template.Template
)

func init() {
	var err error
	TmplMain = template.New("Main")
	TmplMain, err = TmplMain.Parse(tmplMain)
	if err != nil {
		Log.Warnln(err)
	}
}

func (s *State) AsHtml(w io.Writer) {
	body := &HtmlMain{}
	body.Title = "Hello world!"
	m := &HtmlTable{
		Title : "Main info",
		Values : []*KeyValue{
			&KeyValue{"AntEater version", version},
			&KeyValue{"Uptime", fmt.Sprintf("%v", StartTime.Sub(time.Now()))},
			&KeyValue{"Goroutines count", fmt.Sprintf("%d", s.Main.Goroutines)},
			&KeyValue{"Index file size", HumanBytes(s.Main.IndexFileSize)},
		},
	}
	
	f := &HtmlTable{
		Title : "Files",
		Values : []*KeyValue{
			&KeyValue{"Container count", fmt.Sprintf("%d", s.Files.ContainersCount)},
			&KeyValue{"Total files count", fmt.Sprintf("%d", s.Files.FilesCount)},
			&KeyValue{"Avg. Files per container", fmt.Sprintf("%d", SafeDivision(s.Files.FilesCount, int64(s.Files.ContainersCount)))},
			&KeyValue{"Total files size", fmt.Sprintf("%s", HumanBytes(s.Files.FilesSize))},
			&KeyValue{"Avg. file size", fmt.Sprintf("%s", HumanBytes(SafeDivision(s.Files.FilesSize, s.Files.FilesCount)))},
			&KeyValue{"Free spaces count", fmt.Sprintf("%d", s.Files.SpacesCount)},
			&KeyValue{"Free spaces total size", fmt.Sprintf("%s", HumanBytes(s.Files.SpacesSize))},
			&KeyValue{"Avg. free space size", fmt.Sprintf("%s", HumanBytes(SafeDivision(s.Files.SpacesSize, s.Files.SpacesCount)))},
			&KeyValue{"Free spaces percents", fmt.Sprintf("%d %s", SafeDivision(s.Files.SpacesSize * 100, s.Files.FilesSize), "%")},
		},
	}
	
	body.Tables = []*HtmlTable{m, f}
	TmplMain.Execute(w, body)
}

func SafeDivision(a, b int64) int64 {
	if b <= 0 {
		return 0
	}
	return a / b
}

func HumanBytes(size int64) (result string) {
	switch {
		case size > (1024 * 1024 * 1024 * 1024):
			result = fmt.Sprintf("%6.2f TiB", float64(size) / 1024 / 1024 / 1024 / 1024)
		case size > (1024 * 1024 * 1024):
			result = fmt.Sprintf("%6.2f GiB", float64(size) / 1024 / 1024 / 1024)
		case size > (1024 * 1024):
			result = fmt.Sprintf("%6.2f MiB", float64(size) / 1024 / 1024)
		case size > 1024:
			result = fmt.Sprintf("%6.2f KiB", float64(size) / 1024)
		default :
			result = fmt.Sprintf("%d B", size)
	}
	return
}

func Round(val float64, prec int) float64 {
    var rounder float64
    intermed := val * math.Pow(10, float64(prec))

    if val >= 0.5 {
        rounder = math.Ceil(intermed)
    } else {
        rounder = math.Floor(intermed)
    }
    return rounder / math.Pow(10, float64(prec))
}

