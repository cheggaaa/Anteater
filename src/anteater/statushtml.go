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
			table tr tr:nth-child(2n+1) {
			  	background-color: #99ff99;
			}
		
			.container, .container div {
				height:30px;
			}
			.container {
				background:#FFFFCC;
				border:1px solid #666666;
				border-radius:5px;
				width: 95%;
			}
			.data {
				float:left;
				background:#66CC66;
				width: {{.PData}}%;
			}
			.spaces {
				float:left;
				background:#993333;
				width: {{.PSpaces}}%;
			}
			
			.block {
				float:left;
				margin:20px;
				border:1px solid #000;
				border-radius:10px;
				padding:20px
			}
		</style>
	</head>
	<body>
	<h2>{{.Title}}</h2>
	<div class="container">
		<div class="data"></div>
		<div class="spaces"></div>
	</div>
	<div style="clear:both"></div>
	<div>
	{{with .Tables}}
		{{range .}}
		<div class="block">
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
	<div>
	{{with .Rates}}
	<div class="block">
		<h3>{{.Title}}</h3>
		<table>
			<thead>
				<tr>
					<th></th>
					{{range .Heads}}
					<th scope="col">{{.}}</th>
					{{end}}
				</tr>
			</thead>
			<tfoot>
				{{with .Total}}
		        <tr>
		            <th scope="row">.Name</th>
		            {{range .Values}}
					<td scope="col">{{.}}</td>
					{{end}}
		        </tr>
		        {{end}}
		    </tfoot>
		    <tbody>
		    	{{with .Rates}}
			    <tr>
			        <th scope="row">.Name</th>
			        {{range .Values}}
					<td scope="col">{{.}}</td>
					{{end}}
			        </tr>
	        	{{end}}
        	</tbody>
		</table>
	</div>
	{{end}}
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
	Rates  *HtmlRates
	PData int64
	PSpaces int64
}

type HtmlTable struct {
	Title string
	Values []*KeyValue
}

type HtmlRates struct {
	Title string
	Heads []string
	Rates []*HtmlRate
	Total *HtmlRate
}

type HtmlRate struct {
	Name string
	Values []string
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
	
	rates := &HtmlRates{}
	rates.Title = "Request rates"
	rates.Heads = []string{"5 seconds", "1 minute", "5 minutes", "Since start"}
	rates.Rates = []*HtmlRate{
		&HtmlRate{"Get", []string{
			fmt.Sprintf("%d", s.RatesLast5Seconds.Get),
			fmt.Sprintf("%d", s.RatesLastMinute.Get),
			fmt.Sprintf("%d", s.RatesLast5Minutes.Get),
			fmt.Sprintf("%d", s.RatesSinceStart.Get),
		}},
		&HtmlRate{"Add", []string{
			fmt.Sprintf("%d", s.RatesLast5Seconds.Add),
			fmt.Sprintf("%d", s.RatesLastMinute.Add),
			fmt.Sprintf("%d", s.RatesLast5Minutes.Add),
			fmt.Sprintf("%d", s.RatesSinceStart.Add),
		}},
		&HtmlRate{"Delete", []string{
			fmt.Sprintf("%d", s.RatesLast5Seconds.Delete),
			fmt.Sprintf("%d", s.RatesLastMinute.Delete),
			fmt.Sprintf("%d", s.RatesLast5Minutes.Delete),
			fmt.Sprintf("%d", s.RatesSinceStart.Delete),
		}},
		&HtmlRate{"Not Found", []string{
			fmt.Sprintf("%d", s.RatesLast5Seconds.NotFound),
			fmt.Sprintf("%d", s.RatesLastMinute.NotFound),
			fmt.Sprintf("%d", s.RatesLast5Minutes.NotFound),
			fmt.Sprintf("%d", s.RatesSinceStart.NotFound),
		}},
	}
	rates.Total = &HtmlRate{"Total", []string{
			fmt.Sprintf("%d", s.RatesLast5Seconds.Sum()),
			fmt.Sprintf("%d", s.RatesLastMinute.Sum()),
			fmt.Sprintf("%d", s.RatesLast5Minutes.Sum()),
			fmt.Sprintf("%d", s.RatesSinceStart.Sum()),
		}}
		
	body.Tables = []*HtmlTable{m, f, a, c}
	body.Rates = rates
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


