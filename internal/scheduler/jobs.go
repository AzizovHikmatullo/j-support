package scheduler

import (
	"context"
	"time"
)

const (
	ticketTimeout = 10 * time.Second

	inactivityCloseMessage = "Время ожидания ответа истекло. Ваше обращение будет закрыто!"
)

func (sch *Scheduler) registerJobs() {
	// CLOSE TICKET AFTER 5min INACTIVITY

	sch.s.Every(1).Minute().Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		sessions, err := sch.scenarioRepo.GetInactiveSessions(ctx, time.Now().Add(-ticketTimeout))
		if err != nil {
			sch.logger.Error("failed to get inactive sessions", "error", err)
			return
		}

		for _, s := range sessions {
			_, _ = sch.ticketService.CreateMessage(ctx, s.TicketID, 0, "bot", inactivityCloseMessage)
			_ = sch.ticketService.ChangeStatus(ctx, 0, "bot", s.TicketID, "closed")
		}
	})
}
