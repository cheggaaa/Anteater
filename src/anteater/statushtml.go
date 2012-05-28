/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

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
			body {
				background:#fff;
				padding:20px;
			}
			table tr tr:nth-child(2n+1) {
			  	background-color: #99ff99;
			}		
			.container, .container div {
				height:30px;
			}
			.container div {
				float:left;
				border:1px solid #666666;
				border-right:none;
				border-left:none;
			}
			.container div:last-child {
				border-right:1px solid #666666;
				border-top-right-radius:5px;
				border-bottom-right-radius:5px;
				margin-right:-5%;
			}
			.container div:first-child {
				border-left:1px solid #666666;
				border-top-left-radius:5px;
				border-bottom-left-radius:5px;
			}
			.container {
				width: 95%;
				margin-left:20px;
			}
			.data {
				background:#66CC66;		
			}
			.data:nth-child(odd) {
				background: #55BB55;
			}
			.spaces {
				background:#993333;
			}
			.free {
				background:#FFFFCC;
			}
			.block {
				float:left;
				margin:20px;
				border:1px solid #000;
				border-radius:10px;
				padding:20px;
				padding-top:5px;
			}
			
			table.rates {
			    font-style: normal;
			    font-weight: normal;
			    text-align:center;
			    border-collapse:collapse;
			}
			.rates thead th {
			    padding:6px 10px;
			    color:#444;
			    background-color:#FFFFCC;
			    font-weight:bold;
			}
			.rates thead th:empty {
			    background:transparent;
			    border:none;
			}
			.rates tfoot :nth-child(2) {
			    border-bottom-left-radius:5px;
			}
			.rates thead :nth-child(2) {
			    border-top-left-radius:5px;
			}
			
			.rates thead :nth-child(5) {
			    border-top-right-radius:5px;
			}
			.rates tfoot :nth-child(5) {
			    border-bottom-right-radius:5px;
			}
			.rates tfoot td {
			    font-weight:bold;
			    background-color:#FFFFCC;
			}
			.rates tbody td {
				border:1px solid #aaa;  
			}
			.rates tbody td:nth-child(even) {
			    color:#444;		    
			}
			.rates tbody td:nth-child(odd) {
			    background-color:#FFFFCC;
			    color:#000;		    
			}
			.rates tbody th, .rates tfoot th {
			    color:#696969;
			    background-color:#FFFFCC;
			    text-align:right;
			    padding:0px 10px;
			}
			.rates tfoot th {
				color:#000;
			}
		</style>
	</head>
	<body>
	<h2>{{.Title}}</h2>
	<div class="container">
		{{with .Containers}}
		{{range .}}
		<div class="data" style="width:{{.Data}}%"></div>
		<div class="spaces" style="width:{{.Spaces}}%"></div>
		<div class="free" style="width:{{.Free}}%"></div>
		{{end}}
		{{end}}
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
		<table class="rates">
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
		            <th scope="row">{{.Name}}</th>
		            {{range .Values}}
					<td scope="col">{{.}}</td>
					{{end}}
		        </tr>
		        {{end}}
		    </tfoot>
		    <tbody>
		    	{{with .Rates}}
		    	{{range .}}
			    <tr>
			        <th scope="row">{{.Name}}</th>
			        {{range .Values}}
					<td scope="col">{{.}}</td>
					{{end}}
			    </tr>
			    {{end}}
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
	Containers []*HtmlContainer
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

type HtmlContainer struct {
	Data, Spaces, Free float64 
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
	body.Title = "Anteater server status"
	
	allocated := int64(0)
	for _, c := range s.Files.ByContainers {
		allocated += c[0]
	}
	
	cont := make([]*HtmlContainer, s.Files.ContainersCount)
	
	i := 0
	for _, c := range s.Files.ByContainers {
		cont[i] = &HtmlContainer{
			float64(SafeDivision(c[1] * 100000, allocated)) / 1000,
			float64(SafeDivision(c[2] * 100000, allocated)) / 1000,
			float64(SafeDivision((c[0] - c[1]) * 100000, allocated)) / 1000,
		}
		i++
	}
	
	body.Containers = cont
	
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

func DurationToString(d *time.Duration) (result string) {
	return
}


