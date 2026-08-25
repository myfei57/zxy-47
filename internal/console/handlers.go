package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"greenhouse/internal/climate"
	"greenhouse/internal/curtain"
	"greenhouse/internal/env"
	"greenhouse/internal/fert"
	"greenhouse/internal/irrig"
	"greenhouse/internal/ns"
	"greenhouse/internal/quota"
	"greenhouse/internal/shed"
	"greenhouse/internal/task"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func shedID(r *http.Request) string {
	return chi.URLParam(r, "shedID")
}

func (s *Server) apiListSheds(w http.ResponseWriter, r *http.Request) {
	sheds, err := s.services.Sheds.List("")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sheds)
}

func (s *Server) apiCreateShed(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SiteID     string  `json:"site_id"`
		Name       string  `json:"name"`
		Area       float64 `json:"area"`
		Zones      int     `json:"zones"`
		LightZones int     `json:"light_zones"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	siteID := input.SiteID
	if siteID == "" {
		site := ns.NewSite("默认基地", "本地")
		err := s.services.Sites.Save(site)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		siteID = site.ID
	}
	if input.Zones <= 0 {
		input.Zones = 8
	}
	if input.LightZones <= 0 {
		input.LightZones = input.Zones
	}
	created := shed.New(siteID, input.Name, input.Area, input.Zones, input.LightZones)
	if err := s.services.Sheds.Register(created); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.bootstrapShed(created.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) apiShedDetail(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	sh, err := s.services.Sheds.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	partition, _ := s.services.Sheds.LoadPartition(id)
	bindings, _ := s.services.Bindings.Load(id)
	layout := shed.BuildLayout(id, partition.Zones, partition.LightZones)
	firstZone, _ := layout.ZoneAt(1)
	_, firstZoneIndex, _ := ns.ParseZoneID(firstZone.ID)
	siteName := ""
	if site, err := s.services.Sites.Get(sh.SiteID); err == nil {
		siteName = site.DisplayName()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"shed":            sh,
		"summary":         s.services.Sheds.Summary(sh),
		"site":            siteName,
		"store_dir":       s.services.Store.Dir(),
		"partition":       partition,
		"light_zone_count": layout.LightZoneCount(),
		"first_zone":      firstZone,
		"first_zone_name": ns.ZoneName(firstZone.Index),
		"first_zone_index": firstZoneIndex,
		"bindings":        bindings,
	})
}

func (s *Server) apiShedSensors(w http.ResponseWriter, r *http.Request) {
	sensors, err := s.services.Sensors.ListByShed(shedID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	type sensorView struct {
		Sensor env.Sensor `json:"sensor"`
		Zone   string     `json:"zone"`
	}
	out := make([]sensorView, 0, len(sensors))
	for _, sensor := range sensors {
		zone, _ := s.services.Bindings.ZoneOf(shedID(r), sensor.ID)
		out = append(out, sensorView{Sensor: sensor, Zone: zone})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) apiRegisterSensor(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ZoneID string        `json:"zone_id"`
		Type   env.SensorType `json:"type"`
		Name   string        `json:"name"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sensor := env.NewSensor(shedID(r), input.ZoneID, input.Type, input.Name)
	if err := s.services.Sensors.Register(sensor); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = s.services.Bindings.RefreshBindings(shedID(r), s.zoneIDs(shedID(r)))
	writeJSON(w, http.StatusCreated, sensor)
}

func (s *Server) apiClimate(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	mode, _ := s.services.Climate.Current(id)
	guards := s.services.Cool.Guards(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":          mode,
		"guards":        guards,
		"filter_window": s.services.Fans.FilterWindow(),
	})
}

func (s *Server) apiSetMode(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Mode string `json:"mode"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	state, err := s.services.Climate.Switch(shedID(r), input.Mode)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) apiCool(w http.ResponseWriter, r *http.Request) {
	if err := s.services.Cool.Cool(shedID(r)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cooled"})
}

func (s *Server) apiEvaluate(w http.ResponseWriter, r *http.Request) {
	if err := s.services.Monitor.Evaluate(shedID(r)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

func (s *Server) apiCurtain(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	zone := id + "-Z01"
	position, err := s.services.Curtain.Position(id, zone)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	lamp, _ := s.services.Curtain.Lamp(id, zone)
	writeJSON(w, http.StatusOK, map[string]any{
		"position":   position,
		"lamp":       lamp,
		"sunlight":   s.services.Sun.Value(id, 0),
		"thresholds": s.services.Curtain.Thresholds(id),
	})
}

func (s *Server) apiRetract(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	if err := s.services.Curtain.Retract(id, id+"-Z01"); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "retracted"})
}

func (s *Server) apiExtend(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	if err := s.services.Curtain.Extend(id, id+"-Z01"); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "extended"})
}

func (s *Server) apiVents(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	vents, err := s.services.Vent.Sequence(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	list, err := s.services.Vent.List(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"vents": list, "sequence": vents})
}

func (s *Server) apiFilm(w http.ResponseWriter, r *http.Request) {
	films, err := s.services.Roller.Films(shedID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, films)
}

func (s *Server) apiFilmCommands(w http.ResponseWriter, r *http.Request) {
	marks, err := s.services.Roller.ExecutedCommands()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, marks)
}

func (s *Server) apiFilmReset(w http.ResponseWriter, r *http.Request) {
	if err := s.services.Roller.ResetCommands(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func (s *Server) apiFilmRoll(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CmdID  string `json:"cmd_id"`
		FilmID string `json:"film_id"`
		Action string `json:"action"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	position, err := s.services.Roller.Roll(shedID(r), input.CmdID, input.FilmID, input.Action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"position": position})
}

func (s *Server) apiFilmRetry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CmdID  string `json:"cmd_id"`
		FilmID string `json:"film_id"`
		Action string `json:"action"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	position, err := s.services.Roller.Retry(shedID(r), input.CmdID, input.FilmID, input.Action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"position": position})
}

func (s *Server) apiFert(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	recipe := s.services.Fert.Recipe(id)
	mix, _ := s.services.Mixer.Latest(id)
	ph, _ := s.services.PH.Latest(id)
	ec, _ := s.services.ECIrrig.Current(id)
	dose, _ := s.services.Doses.Last(id)
	writeJSON(w, http.StatusOK, map[string]any{"recipe": recipe, "mix": mix, "ph": ph, "ec": ec, "dose": dose})
}

func (s *Server) apiFertAdjust(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PH float64 `json:"ph"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.services.Fert.Adjust(shedID(r), input.PH)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) apiFertRecipe(w http.ResponseWriter, r *http.Request) {
	var input fert.Recipe
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.ShedID = shedID(r)
	if err := s.services.Fert.SetRecipe(shedID(r), input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, input)
}

func (s *Server) apiPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := s.services.Plans.List(shedID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

func (s *Server) apiCreatePlan(w http.ResponseWriter, r *http.Request) {
	var input irrig.Plan
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.ShedID = shedID(r)
	if input.ID == "" {
		input.ID = "plan-" + input.ShedID + "-" + randHex(4)
	}
	if err := s.services.Plans.Save(input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, input)
}

func (s *Server) apiRunCycle(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlanID string `json:"plan_id"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.services.Cycle.Run(shedID(r), input.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) apiQuota(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	quotaState, err := s.services.Quotas.Current(id)
	if err != nil {
		quotaState = quota.Quota{ShedID: id}
	}
	allowance, _ := s.services.Allow.Current(id)
	plan := s.services.QuotaPlan.LoadOrDefault(id)
	ledgerSum, _ := s.services.Ledger.Sum(id, quotaState.Cycle)
	writeJSON(w, http.StatusOK, map[string]any{"quota": quotaState, "allowance": allowance, "plan": plan, "ledger_sum": ledgerSum})
}

func (s *Server) apiQuotaRelease(w http.ResponseWriter, r *http.Request) {
	state, err := s.services.Quotas.Release(shedID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) apiTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.services.Tasks.List(shedID(r), true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) apiCreateTask(w http.ResponseWriter, r *http.Request) {
	var input task.Task
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.ShedID = shedID(r)
	created, err := s.services.Tasks.Create(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) apiCompleteTask(w http.ResponseWriter, r *http.Request) {
	if err := s.services.Tasks.Complete(chi.URLParam(r, "taskID")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (s *Server) apiStages(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	plan := s.services.StagePlan.LoadOrDefault(id)
	current, _ := s.services.Stages.Current(id)
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan, "current": current})
}

func (s *Server) apiSwitchStage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Label string `json:"label"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	template, ok := s.services.StagePlan.Template(shedID(r), input.Label)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown stage label")
		return
	}
	stage := task.Stage{ShedID: shedID(r), Name: input.Label, Label: input.Label, EC: template.EC, PH: template.PH}
	if err := s.services.Stages.Switch(stage); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stage)
}

func (s *Server) apiReadings(w http.ResponseWriter, r *http.Request) {
	id := shedID(r)
	out := map[string]any{}
	for _, kind := range []env.SensorType{env.TypeTemperature, env.TypeHumidity, env.TypeSunlight, env.TypeMoisture} {
		reading, err := s.services.Sampler.Latest(id, kind)
		if err == nil {
			out[string(kind)] = reading
		}
	}
	out["sunlight_durable"] = s.services.Sun.Value(id, 0)
	out["moisture_durable"] = s.services.Moisture.Value(id, 0)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) apiSetReading(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type  env.SensorType `json:"type"`
		Value float64        `json:"value"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	reading, err := s.services.Sampler.SetReading(shedID(r), input.Type, input.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if input.Type == env.TypeTemperature {
		if err := s.services.Sampler.AppendTemp(shedID(r), input.Value); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, reading)
}

func (s *Server) apiAudit(w http.ResponseWriter, r *http.Request) {
	rows, err := s.services.Audit.Recent(shedID(r), 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	count, _ := s.services.Audit.CountByAction(shedID(r), "run")
	writeJSON(w, http.StatusOK, map[string]any{"records": rows, "run_count": count})
}

func (s *Server) apiAlarms(w http.ResponseWriter, r *http.Request) {
	rows, err := s.services.Alarms.List("", 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) apiAckAlarm(w http.ResponseWriter, r *http.Request) {
	if err := s.services.Alarms.Ack(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "acked"})
}

func (s *Server) apiRepartition(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Zones      int `json:"zones"`
		LightZones int `json:"light_zones"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	layout, err := s.services.Sheds.RePartition(shedID(r), input.Zones, input.LightZones)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, layout)
}

func (s *Server) apiSetGuards(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Guards      climate.GuardParams `json:"guards"`
		FanUpper    float64             `json:"fan_upper"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.services.Cool.SetGuards(shedID(r), input.Guards); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if input.FanUpper > 0 {
		if err := s.services.Fans.SetUpperBound(shedID(r), input.FanUpper); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "guards-saved"})
}

func (s *Server) apiSetThresholds(w http.ResponseWriter, r *http.Request) {
	var input curtain.Thresholds
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.services.Curtain.SetThresholds(shedID(r), input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "thresholds-saved"})
}

func (s *Server) apiCalendar(w http.ResponseWriter, r *http.Request) {
	events, err := s.services.Calendar.List(shedID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	next, err := s.services.Calendar.Next(shedID(r), 0)
	if err != nil {
		next = task.CalendarEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "next": next})
}

func (s *Server) apiAddCalendar(w http.ResponseWriter, r *http.Request) {
	var input task.CalendarEvent
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.ShedID = shedID(r)
	if err := s.services.Calendar.Add(input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, input)
}

func (s *Server) bootstrapShed(id string) error {
	if err := s.services.Vent.Register(id); err != nil {
		return err
	}
	if err := s.services.Fans.Register(id); err != nil {
		return err
	}
	if err := s.services.Roller.Register(id, id+"-film-east", "东边卷膜", 5); err != nil {
		return err
	}
	if err := s.services.Roller.Register(id, id+"-film-west", "西边卷膜", 5); err != nil {
		return err
	}
	if _, err := s.services.Allow.Reset(id, 100); err != nil {
		return err
	}
	if err := s.services.QuotaPlan.Save(quota.DefaultPlan(id)); err != nil {
		return err
	}
	zone := id + "-Z01"
	for _, seed := range []struct {
		kind env.SensorType
		name string
	}{
		{env.TypeTemperature, "温湿度探头"},
		{env.TypeHumidity, "湿度探头"},
		{env.TypeSunlight, "光照探头"},
		{env.TypeMoisture, "墒情探头"},
	} {
		if err := s.services.Sensors.Register(env.NewSensor(id, zone, seed.kind, seed.name)); err != nil {
			return err
		}
	}
	return s.services.Bindings.RefreshBindings(id, s.zoneIDs(id))
}

func (s *Server) zoneIDs(id string) []string {
	partition, err := s.services.Sheds.LoadPartition(id)
	if err != nil {
		return []string{id + "-Z01"}
	}
	layout := shed.BuildLayout(id, partition.Zones, partition.LightZones)
	return layout.ZoneIDs()
}

func randHex(length int) string {
	const digits = "0123456789abcdef"
	out := make([]byte, length)
	value := timeNowUnixNano()
	for i := 0; i < length; i++ {
		out[i] = digits[value%16]
		value /= 16
	}
	return string(out)
}

func timeNowUnixNano() int64 {
	return unixNano()
}
