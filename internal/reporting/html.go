package reporting

import (
	"fmt"
	"html/template"
	"io"

	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type htmlSuiteReporter struct{}

func (htmlSuiteReporter) Write(writer io.Writer, suite domain.ScenarioSuiteResult) error {
	if err := suiteHTML.Execute(writer, suite); err != nil {
		return fmt.Errorf("render HTML report: %w", err)
	}
	return nil
}

var suiteHTML = template.Must(template.New("mobilelab-report").Funcs(template.FuncMap{
	"status": func(passed bool) string {
		if passed {
			return "PASS"
		}
		return "FAIL"
	},
	"checks": func(result domain.ScenarioResult) []domain.ScenarioCheck {
		return append(append([]domain.ScenarioCheck(nil), result.Steps...), result.Assertions...)
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>MobileLab report — {{.Name}}</title>
  <style>
    :root{color-scheme:light dark;--bg:#f5f7fb;--card:#fff;--text:#172033;--muted:#667085;--pass:#067647;--fail:#b42318;--line:#e4e7ec}
    @media(prefers-color-scheme:dark){:root{--bg:#101828;--card:#1d2939;--text:#f2f4f7;--muted:#98a2b3;--line:#344054}}
    *{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px/1.5 system-ui,-apple-system,sans-serif}main{max-width:1040px;margin:0 auto;padding:32px 20px}h1{margin:0 0 4px;font-size:28px}.meta{color:var(--muted);margin:0 0 24px}.summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin-bottom:24px}.metric,.scenario{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:16px}.metric strong{display:block;font-size:24px}.pass{color:var(--pass)}.fail{color:var(--fail)}.scenario{margin-bottom:14px}.scenario h2{font-size:18px;margin:0}.scenario header{display:flex;justify-content:space-between;gap:12px}.duration{color:var(--muted)}table{width:100%;border-collapse:collapse;margin-top:12px}th,td{text-align:left;padding:8px;border-top:1px solid var(--line);vertical-align:top}th{color:var(--muted);font-weight:600}.message{white-space:pre-wrap}.error{border-left:3px solid var(--fail);padding-left:10px}@media(max-width:620px){.summary{grid-template-columns:repeat(2,1fr)}}
  </style>
</head>
<body><main>
  <h1>{{.Name}}</h1>
  <p class="meta">Started {{.StartedAt.UTC.Format "2006-01-02 15:04:05Z07:00"}} · MobileLab CI report</p>
  <section class="summary" aria-label="Suite summary">
    <div class="metric"><span>Status</span><strong class="{{if .Passed}}pass{{else}}fail{{end}}">{{status .Passed}}</strong></div>
    <div class="metric"><span>Scenarios</span><strong>{{.Summary.Total}}</strong></div>
    <div class="metric"><span>Passed</span><strong class="pass">{{.Summary.Passed}}</strong></div>
    <div class="metric"><span>Failed</span><strong class="fail">{{.Summary.Failed}}</strong></div>
  </section>
  {{range .Scenarios}}
  <article class="scenario">
    <header><h2>{{.Name}}</h2><strong class="{{if .Passed}}pass{{else}}fail{{end}}">{{status .Passed}}</strong></header>
    <div class="duration">{{.DurationMS}}ms</div>
    {{if .Error}}<p class="error">{{.Error}}</p>{{end}}
    <table><thead><tr><th>Check</th><th>Result</th><th>Message</th></tr></thead><tbody>
      {{range checks .}}<tr><td>{{.Name}}</td><td class="{{if .Passed}}pass{{else}}fail{{end}}">{{status .Passed}}</td><td class="message">{{.Message}}</td></tr>{{else}}<tr><td>Scenario execution</td><td class="{{if .Passed}}pass{{else}}fail{{end}}">{{status .Passed}}</td><td></td></tr>{{end}}
    </tbody></table>
  </article>
  {{end}}
</main></body></html>
`))
