package anteater

import (
	"html/template"
	"fmt"
	"io"
	"time"
)


const (
	tmplMain = `
<html>
	<head>
		<title>{{.Title}}</title>
		<style type="text/css">
			.container, .container div {
				height:100px;
			}
			.container {
				background:#ddd;
				border:1px solid #bbb;
				border-radius:5px;
				width: 90%;
			}
			.data {
				float:left;
				background:#0f0;
				width: {{.PData}}%;
			}
			.spaces {
				float:left;
				background:#f00;
				width: {{.PSpaces}}%;
			}
		</style>
	</head>
	<body>
	<h2>{{.Title}}</h2>
	<div>
	{{with .Tables}}
		{{range .}}
		<div style="float:left;margin:20px;border:1px solid #000;border-radius:10px;padding:20px">
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
	</div>
	<div style="clear:both"></div>
	<div class="container">
		<div class="data"></div>
		<div class="spaces"></div>
	</div>
	</body>
</html>`

)

type KeyValue struct {
	Key, Value string
}

type HtmlMain struct {
	Title string
	Tables []*HtmlTable
	PData int64
	PSpaces int64
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
	body.Title = "Server status " + version
	
	allocated := Conf.ContainerSize * int64(s.Files.ContainersCount)
	body.PSpaces = SafeDivision(s.Files.SpacesSize * 100, allocated)
	body.PData = SafeDivision(s.Files.FilesSize * 100, allocated)
	
	// round time to second
	dt := time.Unix(LastDump.Unix(), 0)
	st := time.Unix(StartTime.Unix(), 0)
	nt := time.Unix(time.Now().Unix(), 0)
	
	m := &HtmlTable{
		Title : "Main info",
		Values : []*KeyValue{
			&KeyValue{"AntEater version", version},
			&KeyValue{"Uptime", fmt.Sprintf("%v", nt.Sub(st))},
			&KeyValue{"Goroutines count", fmt.Sprintf("%d", s.Main.Goroutines)},
			&KeyValue{"Dump file size", HumanBytes(s.Main.IndexFileSize)},
			&KeyValue{"Last dump", fmt.Sprintf("%v ago, for %v", nt.Sub(dt), LastDumpTime)},
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
	
	c := &HtmlTable{
		Title : "Counters (since start)",
		Values : []*KeyValue{
			&KeyValue{"Total", fmt.Sprintf("%d", s.Counters.Sum())},
			&KeyValue{"Get", fmt.Sprintf("%d", s.Counters.Get)},
			&KeyValue{"Add", fmt.Sprintf("%d", s.Counters.Add)},
			&KeyValue{"Delete", fmt.Sprintf("%d", s.Counters.Delete)},
			&KeyValue{"Not Found", fmt.Sprintf("%d", s.Counters.NotFound)},
		},
	}
	
	ta := s.Alloc.ReplaceSpace + s.Alloc.ToEnd + s.Alloc.AddToSpace
	
	a := &HtmlTable{
		Title : "Allocates",
		Values : []*KeyValue{
			&KeyValue{"Append", fmt.Sprintf("%d %s  (%d)", SafeDivision(s.Alloc.ToEnd * 100, ta), "%", s.Alloc.ToEnd)},
			&KeyValue{"Append to space", fmt.Sprintf("%d %s  (%d)", SafeDivision(s.Alloc.AddToSpace * 100, ta), "%", s.Alloc.AddToSpace)},
			&KeyValue{"Replace space", fmt.Sprintf("%d %s  (%d)", SafeDivision(s.Alloc.ReplaceSpace * 100, ta), "%", s.Alloc.ReplaceSpace)},
		},
	}
	
	rs := &HtmlTable{
		Title : "Rates (since start)",
		Values : []*KeyValue{
			&KeyValue{"Total", fmt.Sprintf("%d rps", s.RatesSinceStart.Sum())},
			&KeyValue{"Get", fmt.Sprintf("%d rps", s.RatesSinceStart.Get)},
			&KeyValue{"Add", fmt.Sprintf("%d rps", s.RatesSinceStart.Add)},
			&KeyValue{"Delete", fmt.Sprintf("%d rps", s.RatesSinceStart.Delete)},
			&KeyValue{"Not Found", fmt.Sprintf("%d rps", s.RatesSinceStart.NotFound)},
		},
	}
	
	r5m := &HtmlTable{
		Title : "Rates (5 minutes)",
		Values : []*KeyValue{
			&KeyValue{"Total", fmt.Sprintf("%d rps", s.RatesLast5Minutes.Sum())},
			&KeyValue{"Get", fmt.Sprintf("%d rps", s.RatesLast5Minutes.Get)},
			&KeyValue{"Add", fmt.Sprintf("%d rps", s.RatesLast5Minutes.Add)},
			&KeyValue{"Delete", fmt.Sprintf("%d rps", s.RatesLast5Minutes.Delete)},
			&KeyValue{"Not Found", fmt.Sprintf("%d rps", s.RatesLast5Minutes.NotFound)},
		},
	}
	
	r1m := &HtmlTable{
		Title : "Rates (1 minute)",
		Values : []*KeyValue{
			&KeyValue{"Total", fmt.Sprintf("%d rps", s.RatesLastMinute.Sum())},
			&KeyValue{"Get", fmt.Sprintf("%d rps", s.RatesLastMinute.Get)},
			&KeyValue{"Add", fmt.Sprintf("%d rps", s.RatesLastMinute.Add)},
			&KeyValue{"Delete", fmt.Sprintf("%d rps", s.RatesLastMinute.Delete)},
			&KeyValue{"Not Found", fmt.Sprintf("%d rps", s.RatesLastMinute.NotFound)},
		},
	}
	
	r5s := &HtmlTable{
		Title : "Rates (5 seconds)",
		Values : []*KeyValue{
			&KeyValue{"Total", fmt.Sprintf("%d p/s", s.RatesLast5Seconds.Sum())},
			&KeyValue{"Get", fmt.Sprintf("%d p/s", s.RatesLast5Seconds.Get)},
			&KeyValue{"Add", fmt.Sprintf("%d p/s", s.RatesLast5Seconds.Add)},
			&KeyValue{"Delete", fmt.Sprintf("%d p/s", s.RatesLast5Seconds.Delete)},
			&KeyValue{"Not Found", fmt.Sprintf("%d p/s", s.RatesLast5Seconds.NotFound)},
		},
	}
	
	body.Tables = []*HtmlTable{m, f, a, c, rs, r5m, r1m, r5s}
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


