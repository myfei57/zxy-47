package console

import (
	"html/template"
	"net/http"
	"time"
)

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))
var shedsTemplate = template.Must(template.New("sheds").Parse(shedsHTML))
var climateTemplate = template.Must(template.New("climate").Parse(climateHTML))
var irrigTemplate = template.Must(template.New("irrig").Parse(irrigHTML))
var alarmsTemplate = template.Must(template.New("alarms").Parse(alarmsHTML))

type pageBase struct {
	Title string
	Nav   []navLink
}

type navLink struct {
	Href string
	Text string
}

func base(title string) pageBase {
	return pageBase{
		Title: title,
		Nav: []navLink{
			{Href: "/sheds", Text: "大棚"},
			{Href: "/climate", Text: "气候"},
			{Href: "/irrig", Text: "灌溉"},
			{Href: "/alarms", Text: "告警"},
		},
	}
}

type shedRow struct {
	Name      string
	Zones     int
	LightZones int
	Area      float64
	Created   string
}

type climateRow struct {
	Name  string
	Mode  string
	Wind  string
	Upper float64
	Sun   float64
	Curtain string
	Lamp  string
}

type planRow struct {
	Name      string
	Threshold float64
	Baseline  float64
	Enabled   string
}

type irrigRow struct {
	Shed     string
	Plans    []planRow
	Allowance float64
	Remain   float64
	Cycle    int
}

type alarmRow struct {
	Severity string
	Shed     string
	Message  string
	Acked    string
	At       string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	sheds, _ := s.services.Sheds.List("")
	alarms, _ := s.services.Alarms.List("", 0)
	unacked := 0
	for _, alarm := range alarms {
		if !alarm.Acked {
			unacked++
		}
	}
	_ = s.render(w, indexTemplate, map[string]any{
		"Base":     base("GreenHouse 温室群控平台"),
		"Sheds":    len(sheds),
		"Unacked":  unacked,
	})
}

func (s *Server) handleShedsPage(w http.ResponseWriter, r *http.Request) {
	sheds, err := s.services.Sheds.List("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows := make([]shedRow, 0, len(sheds))
	for _, sh := range sheds {
		rows = append(rows, shedRow{
			Name:       sh.Name,
			Zones:      sh.Zones,
			LightZones: sh.EffectiveLightZones(),
			Area:       sh.Area,
			Created:    sh.CreatedAt.Format(time.RFC3339),
		})
	}
	_ = s.render(w, shedsTemplate, map[string]any{"Base": base("大棚分区 - GreenHouse"), "Rows": rows})
}

func (s *Server) handleClimatePage(w http.ResponseWriter, r *http.Request) {
	sheds, err := s.services.Sheds.List("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows := make([]climateRow, 0, len(sheds))
	for _, sh := range sheds {
		mode, _ := s.services.Climate.Current(sh.ID)
		guards := s.services.Cool.Guards(sh.ID)
		position, _ := s.services.Curtain.Position(sh.ID, sh.ID+"-Z01")
		lamp, _ := s.services.Curtain.Lamp(sh.ID, sh.ID+"-Z01")
		curtainState := "unknown"
		if position.ZoneID != "" {
			curtainState = position.Position
		}
		lampState := "off"
		if lamp.On {
			lampState = "on"
		}
		rows = append(rows, climateRow{
			Name:    sh.Name,
			Mode:    mode.Mode,
			Wind:    guards.Wind,
			Upper:   guards.TempHigh,
			Sun:     guards.SunHigh,
			Curtain: curtainState,
			Lamp:    lampState,
		})
	}
	_ = s.render(w, climateTemplate, map[string]any{"Base": base("气候控制 - GreenHouse"), "Rows": rows})
}

func (s *Server) handleIrrigPage(w http.ResponseWriter, r *http.Request) {
	sheds, err := s.services.Sheds.List("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows := make([]irrigRow, 0, len(sheds))
	for _, sh := range sheds {
		plans, _ := s.services.Plans.List(sh.ID)
		planRows := make([]planRow, 0, len(plans))
		for _, plan := range plans {
			enabled := "停用"
			if plan.Enabled {
				enabled = "启用"
			}
			planRows = append(planRows, planRow{
				Name:      plan.Name,
				Threshold: plan.Threshold,
				Baseline:  plan.BaselineMoisture,
				Enabled:   enabled,
			})
		}
		allowance, _ := s.services.Allow.Current(sh.ID)
		quotaState, _ := s.services.Quotas.Current(sh.ID)
		rows = append(rows, irrigRow{
			Shed:      sh.Name,
			Plans:     planRows,
			Allowance: allowance.Total,
			Remain:    allowance.Remain,
			Cycle:     quotaState.Cycle,
		})
	}
	_ = s.render(w, irrigTemplate, map[string]any{"Base": base("灌溉计划 - GreenHouse"), "Rows": rows})
}

func (s *Server) handleAlarmsPage(w http.ResponseWriter, r *http.Request) {
	alarms, err := s.services.Alarms.List("", 200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows := make([]alarmRow, 0, len(alarms))
	for _, alarm := range alarms {
		acked := "未确认"
		if alarm.Acked {
			acked = "已确认"
		}
		rows = append(rows, alarmRow{
			Severity: alarm.Severity,
			Shed:     alarm.ShedID,
			Message:  alarm.Message,
			Acked:    acked,
			At:       alarm.At.Format(time.RFC3339),
		})
	}
	_ = s.render(w, alarmsTemplate, map[string]any{"Base": base("告警中心 - GreenHouse"), "Rows": rows})
}

func (s *Server) render(w http.ResponseWriter, tmpl *template.Template, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{{.Base.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;background:#f5f7f6;color:#24302b}
nav a{margin-right:1rem;color:#1b7a4b;text-decoration:none;font-weight:600}
h1{color:#174a33}
.cards{display:flex;gap:1rem;margin-top:1rem}
.card{background:#fff;border-radius:8px;padding:1rem 1.5rem;box-shadow:0 1px 3px rgba(0,0,0,.12)}
.card b{font-size:2rem;color:#1b7a4b}
</style>
</head>
<body>
<nav>
{{range .Base.Nav}}<a href="{{.Href}}">{{.Text}}</a>{{end}}
</nav>
<h1>GreenHouse 温室大棚环境群控平台</h1>
<div class="cards">
<div class="card"><div>在管大棚</div><b>{{.Sheds}}</b></div>
<div class="card"><div>未确认告警</div><b>{{.Unacked}}</b></div>
</div>
</body>
</html>`

const shedsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{{.Base.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;background:#f5f7f6;color:#24302b}
nav a{margin-right:1rem;color:#1b7a4b;text-decoration:none;font-weight:600}
table{border-collapse:collapse;background:#fff;width:100%}
th,td{border:1px solid #dce3df;padding:.5rem .75rem;text-align:left}
th{background:#eaf3ee}
</style>
</head>
<body>
<nav>
{{range .Base.Nav}}<a href="{{.Href}}">{{.Text}}</a>{{end}}
</nav>
<h1>大棚分区</h1>
<table>
<tr><th>名称</th><th>分区数</th><th>补光区段</th><th>面积</th><th>创建时间</th></tr>
{{range .Rows}}
<tr><td>{{.Name}}</td><td>{{.Zones}}</td><td>{{.LightZones}}</td><td>{{.Area}}</td><td>{{.Created}}</td></tr>
{{end}}
</table>
</body>
</html>`

const climateHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{{.Base.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;background:#f5f7f6;color:#24302b}
nav a{margin-right:1rem;color:#1b7a4b;text-decoration:none;font-weight:600}
table{border-collapse:collapse;background:#fff;width:100%}
th,td{border:1px solid #dce3df;padding:.5rem .75rem;text-align:left}
th{background:#eaf3ee}
</style>
</head>
<body>
<nav>
{{range .Base.Nav}}<a href="{{.Href}}">{{.Text}}</a>{{end}}
</nav>
<h1>气候模式与守卫</h1>
<table>
<tr><th>大棚</th><th>模式</th><th>风向</th><th>高温阈值</th><th>光照阈值</th><th>遮阳帘</th><th>补光灯</th></tr>
{{range .Rows}}
<tr><td>{{.Name}}</td><td>{{.Mode}}</td><td>{{.Wind}}</td><td>{{.Upper}}</td><td>{{.Sun}}</td><td>{{.Curtain}}</td><td>{{.Lamp}}</td></tr>
{{end}}
</table>
</body>
</html>`

const irrigHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{{.Base.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;background:#f5f7f6;color:#24302b}
nav a{margin-right:1rem;color:#1b7a4b;text-decoration:none;font-weight:600}
table{border-collapse:collapse;background:#fff;width:100%}
th,td{border:1px solid #dce3df;padding:.5rem .75rem;text-align:left}
th{background:#eaf3ee}
</style>
</head>
<body>
<nav>
{{range .Base.Nav}}<a href="{{.Href}}">{{.Text}}</a>{{end}}
</nav>
<h1>灌溉计划与配额</h1>
{{range .Rows}}
<h2>{{.Shed}}</h2>
<table>
<tr><th>计划</th><th>灌溉阈值</th><th>基准墒情</th><th>状态</th></tr>
{{range .Plans}}
<tr><td>{{.Name}}</td><td>{{.Threshold}}</td><td>{{.Baseline}}</td><td>{{.Enabled}}</td></tr>
{{end}}
</table>
<p>周期 {{.Cycle}}：额度 {{.Allowance}}，剩余 {{.Remain}}</p>
{{end}}
</body>
</html>`

const alarmsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{{.Base.Title}}</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;background:#f5f7f6;color:#24302b}
nav a{margin-right:1rem;color:#1b7a4b;text-decoration:none;font-weight:600}
table{border-collapse:collapse;background:#fff;width:100%}
th,td{border:1px solid #dce3df;padding:.5rem .75rem;text-align:left}
th{background:#eaf3ee}
</style>
</head>
<body>
<nav>
{{range .Base.Nav}}<a href="{{.Href}}">{{.Text}}</a>{{end}}
</nav>
<h1>告警中心</h1>
<table>
<tr><th>级别</th><th>大棚</th><th>内容</th><th>状态</th><th>时间</th></tr>
{{range .Rows}}
<tr><td>{{.Severity}}</td><td>{{.Shed}}</td><td>{{.Message}}</td><td>{{.Acked}}</td><td>{{.At}}</td></tr>
{{end}}
</table>
</body>
</html>`
