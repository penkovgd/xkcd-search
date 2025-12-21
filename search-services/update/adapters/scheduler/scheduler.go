package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"yadro.com/course/update/core"
)

type Scheduler struct {
	cron     gocron.Scheduler
	schedule string
	log      *slog.Logger
	service  core.Updater
}

func New(log *slog.Logger, service core.Updater, schedule string) (*Scheduler, error) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, fmt.Errorf("load UTC location: %w", err)
	}
	sch, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		cron:     sch,
		log:      log,
		service:  service,
		schedule: schedule,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	_, err := s.cron.NewJob(
		gocron.CronJob(s.schedule, false),
		gocron.NewTask(func() {
			err := s.service.Update(ctx)
			if err != nil {
				s.log.Error("scheduled update", "error", err)
			} else {
				s.log.Info("scheduled update completed successfully")
			}
		}),
	)
	if err != nil {
		return fmt.Errorf("cron create job: %w", err)
	}
	s.log.Debug("starting scheduler")
	s.cron.Start()
	return nil
}

func (s *Scheduler) Close() error {
	s.log.Debug("stopping scheduler")
	return s.cron.Shutdown()
}
