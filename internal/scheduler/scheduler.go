package scheduler

import (
	"log/slog"
	"time"

	"github.com/AzizovHikmatullo/j-support/internal/scenario"
	"github.com/AzizovHikmatullo/j-support/internal/tickets"
	"github.com/go-co-op/gocron"
)

type Scheduler struct {
	s      *gocron.Scheduler
	logger *slog.Logger

	ticketService tickets.Service
	scenarioRepo  scenario.Repository
}

func New(ticketService tickets.Service, scenarioRepo scenario.Repository, logger *slog.Logger) *Scheduler {
	sched := gocron.NewScheduler(time.UTC)

	return &Scheduler{
		s:             sched,
		logger:        logger,
		ticketService: ticketService,
		scenarioRepo:  scenarioRepo,
	}
}

func (sch *Scheduler) Start() {
	sch.registerJobs()
	sch.s.StartAsync()
}
