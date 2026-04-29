package scheduler

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"bank-api/internal/integration"
	"bank-api/internal/service"
)

type PaymentScheduler struct {
	creditService *service.CreditService
	userRepo      service.UserRepository
	emailSender   integration.EmailSender
	interval      time.Duration
	logger        *logrus.Logger
}

// NewPaymentScheduler создаёт планировщик с заданным интервалом проверки.
func NewPaymentScheduler(cs *service.CreditService, ur service.UserRepository, es integration.EmailSender, interval time.Duration, l *logrus.Logger) *PaymentScheduler {
	return &PaymentScheduler{
		creditService: cs,
		userRepo:      ur,
		emailSender:   es,
		interval:      interval,
		logger:        l,
	}
}

func (s *PaymentScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.logger.Info("Processing overdue credit payments...")
			notifs, err := s.creditService.ProcessOverdue(context.Background())
			if err != nil {
				s.logger.WithError(err).Error("Overdue processing failed")
				continue
			}
			for userID, msgs := range notifs {
				user, err := s.userRepo.GetByID(context.Background(), userID)
				if err != nil || user == nil {
					s.logger.WithField("userID", userID).Warn("User not found for notification")
					continue
				}
				body := ""
				for _, msg := range msgs {
					body += msg + "<br>"
				}
				if err := s.emailSender.Send(user.Email, "Credit Payment Notification", body); err != nil {
					s.logger.WithError(err).WithField("email", user.Email).Error("Failed to send email")
				}
			}
		}
	}
}
