package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"greenhouse/internal/health"
)

type Server struct {
	services *Services
	version  string
	handler  http.Handler
}

func NewServer(services *Services, version string) *Server {
	server := &Server{services: services, version: version}
	server.handler = server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", health.NewHandler(s.version).ServeHTTP)
	r.Get("/", s.handleIndex)
	r.Get("/sheds", s.handleShedsPage)
	r.Get("/climate", s.handleClimatePage)
	r.Get("/irrig", s.handleIrrigPage)
	r.Get("/alarms", s.handleAlarmsPage)
	r.Route("/api", func(api chi.Router) {
		api.Get("/sheds", s.apiListSheds)
		api.Post("/sheds", s.apiCreateShed)
		api.Get("/sheds/{shedID}", s.apiShedDetail)
		api.Get("/sheds/{shedID}/sensors", s.apiShedSensors)
		api.Post("/sheds/{shedID}/sensors", s.apiRegisterSensor)
		api.Get("/sheds/{shedID}/climate", s.apiClimate)
		api.Post("/sheds/{shedID}/climate/mode", s.apiSetMode)
		api.Post("/sheds/{shedID}/climate/cool", s.apiCool)
		api.Post("/sheds/{shedID}/climate/evaluate", s.apiEvaluate)
		api.Get("/sheds/{shedID}/curtain", s.apiCurtain)
		api.Post("/sheds/{shedID}/curtain/retract", s.apiRetract)
		api.Post("/sheds/{shedID}/curtain/extend", s.apiExtend)
		api.Get("/sheds/{shedID}/vents", s.apiVents)
		api.Get("/sheds/{shedID}/film", s.apiFilm)
		api.Post("/sheds/{shedID}/film/roll", s.apiFilmRoll)
		api.Post("/sheds/{shedID}/film/retry", s.apiFilmRetry)
		api.Get("/sheds/{shedID}/fert", s.apiFert)
		api.Post("/sheds/{shedID}/fert/adjust", s.apiFertAdjust)
		api.Post("/sheds/{shedID}/fert/recipe", s.apiFertRecipe)
		api.Get("/sheds/{shedID}/irrig/plans", s.apiPlans)
		api.Post("/sheds/{shedID}/irrig/plans", s.apiCreatePlan)
		api.Post("/sheds/{shedID}/irrig/cycle", s.apiRunCycle)
		api.Get("/sheds/{shedID}/quota", s.apiQuota)
		api.Post("/sheds/{shedID}/quota/release", s.apiQuotaRelease)
		api.Get("/sheds/{shedID}/tasks", s.apiTasks)
		api.Post("/sheds/{shedID}/tasks", s.apiCreateTask)
		api.Post("/tasks/{taskID}/complete", s.apiCompleteTask)
		api.Get("/sheds/{shedID}/stages", s.apiStages)
		api.Post("/sheds/{shedID}/stages", s.apiSwitchStage)
		api.Get("/sheds/{shedID}/readings", s.apiReadings)
		api.Post("/sheds/{shedID}/readings", s.apiSetReading)
		api.Get("/sheds/{shedID}/audit", s.apiAudit)
		api.Get("/alarms", s.apiAlarms)
		api.Post("/alarms/{id}/ack", s.apiAckAlarm)
	})
	return r
}
