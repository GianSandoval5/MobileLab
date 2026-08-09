package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
	"github.com/mobilelab-dev/mobilelab/internal/events"
)

type Server struct {
	Bus      *events.Bus
	Requests domain.RequestRepository
	Runs     domain.ScenarioRunRepository
	State    func() any
}

func (s Server) Page(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	_, _ = writer.Write([]byte(pageHTML))
}

func (s Server) Events(writer http.ResponseWriter, request *http.Request) {
	if s.Bus == nil || s.Requests == nil || s.Runs == nil || s.State == nil {
		http.Error(writer, "dashboard unavailable", http.StatusServiceUnavailable)
		return
	}
	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "dashboard disconnected")
	connection.SetReadLimit(1024)
	ctx := connection.CloseRead(context.Background())
	stream, cancel, err := s.Bus.Subscribe(ctx)
	if err != nil {
		_ = connection.Close(websocket.StatusInternalError, "event bus unavailable")
		return
	}
	defer cancel()

	requests, err := s.Requests.Recent(ctx, 100)
	if err != nil {
		_ = connection.Close(websocket.StatusInternalError, "request history unavailable")
		return
	}
	runs, err := s.Runs.Recent(ctx, 20)
	if err != nil {
		_ = connection.Close(websocket.StatusInternalError, "scenario history unavailable")
		return
	}
	initial := domain.Event{
		Type: domain.EventSnapshot, Version: 1, Timestamp: time.Now().UTC(),
		Payload: map[string]any{"state": s.State(), "requests": requests, "scenarios": runs},
	}
	if err := writeEvent(ctx, connection, initial); err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case event, open := <-stream:
			if !open {
				return
			}
			writeContext, cancelWrite := context.WithTimeout(ctx, 3*time.Second)
			err := writeEvent(writeContext, connection, event)
			cancelWrite()
			if err != nil {
				return
			}
		}
	}
}

func writeEvent(ctx context.Context, connection *websocket.Conn, event domain.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode dashboard event: %w", err)
	}
	return connection.Write(ctx, websocket.MessageText, data)
}

const pageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>MobileLab Dashboard</title>
  <style>
    :root{color-scheme:light dark;--bg:#0b1020;--panel:#141b2d;--text:#edf2ff;--muted:#9aa8c7;--ok:#45d49d;--warn:#ffca6b;--line:#28334c}
    *{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px ui-monospace,SFMono-Regular,Menlo,monospace}
    main{max-width:1100px;margin:0 auto;padding:32px 20px}h1{font:700 28px system-ui;margin:0}.subtitle{color:var(--muted);margin:8px 0 28px}
    .grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:12px;margin-bottom:24px}.card{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:16px}.label{color:var(--muted);font-size:12px}.value{font:650 20px system-ui;margin-top:7px}.online{color:var(--ok)}
    section{background:var(--panel);border:1px solid var(--line);border-radius:10px;margin-top:16px;overflow:hidden}h2{font:650 16px system-ui;margin:0;padding:16px;border-bottom:1px solid var(--line)}
    table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px 14px;border-bottom:1px solid var(--line)}th{color:var(--muted);font-weight:500}.empty{color:var(--muted);padding:18px}#connection{float:right;font-size:12px;color:var(--warn)}
  </style>
</head>
<body><main>
  <h1>MobileLab <span id="connection">connecting…</span></h1><p class="subtitle">Local mobile scenario environment</p>
  <div class="grid"><div class="card"><div class="label">API</div><div class="value online">ONLINE</div></div><div class="card"><div class="label">LATENCY</div><div class="value" id="latency">0ms</div></div><div class="card"><div class="label">FORCED ERROR</div><div class="value" id="error">inactive</div></div><div class="card"><div class="label">AUTH</div><div class="value" id="auth">active</div></div></div>
  <section><h2>Recent requests</h2><table><thead><tr><th>Time</th><th>Method</th><th>Path</th><th>Status</th><th>Duration</th></tr></thead><tbody id="requests"><tr><td colspan="5" class="empty">Waiting for requests…</td></tr></tbody></table></section>
  <section><h2>Scenario runs</h2><table><thead><tr><th>Time</th><th>Name</th><th>Result</th><th>Duration</th></tr></thead><tbody id="scenarios"><tr><td colspan="4" class="empty">No scenario runs yet.</td></tr></tbody></table></section>
</main><script>
const requestRows=[],scenarioRows=[];
const text=(id,value)=>document.getElementById(id).textContent=value;
function state(value){text('latency',(value.latency_ms||0)+'ms');text('error',value.forced_error?'HTTP '+value.forced_error:'inactive');text('auth',value.auth_expired?'expired':'active')}
function time(value){return new Date(value).toLocaleTimeString()}
function renderRequests(){const body=document.getElementById('requests');body.replaceChildren();requestRows.slice(-100).reverse().forEach(r=>{const row=document.createElement('tr');[time(r.timestamp),r.method,r.path,r.status,r.duration_ms+'ms'].forEach(v=>{const cell=document.createElement('td');cell.textContent=v;row.appendChild(cell)});body.appendChild(row)});if(!requestRows.length)body.innerHTML='<tr><td colspan="5" class="empty">Waiting for requests…</td></tr>'}
function renderScenarios(){const body=document.getElementById('scenarios');body.replaceChildren();scenarioRows.slice(-20).reverse().forEach(r=>{const row=document.createElement('tr');[time(r.started_at),r.name,r.passed?'PASS':'FAIL',r.duration_ms+'ms'].forEach(v=>{const cell=document.createElement('td');cell.textContent=v;row.appendChild(cell)});body.appendChild(row)});if(!scenarioRows.length)body.innerHTML='<tr><td colspan="4" class="empty">No scenario runs yet.</td></tr>'}
function connect(){const socket=new WebSocket((location.protocol==='https:'?'wss://':'ws://')+location.host+'/__mobilelab/events');socket.onopen=()=>text('connection','live');socket.onclose=()=>{text('connection','reconnecting…');setTimeout(connect,1000)};socket.onmessage=message=>{const event=JSON.parse(message.data);if(event.type==='environment.snapshot'){state(event.payload.state);requestRows.push(...(event.payload.requests||[]));scenarioRows.push(...(event.payload.scenarios||[]));renderRequests();renderScenarios()}else if(event.type==='request.recorded'){requestRows.push(event.payload);renderRequests()}else if(event.type==='environment.state_changed'){state(event.payload)}else if(event.type==='scenario.completed'){scenarioRows.push(event.payload);renderScenarios()}}}
connect();
</script></body></html>`
