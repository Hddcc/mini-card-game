package service

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type ActivityAwardConsumer interface {
	Consume(consumerTag string) (<-chan amqp.Delivery, error)
}

func StartActivityAwardConsumer(ctx context.Context, consumer ActivityAwardConsumer, service *ActivityLotteryService, logger *zap.Logger) {
	if consumer == nil || service == nil {
		return
	}
	deliveries, err := consumer.Consume("mini-xiyou-activity-award")
	if err != nil {
		if logger != nil {
			logger.Warn("activity award consumer unavailable", zap.Error(err))
		}
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-deliveries:
				if !ok {
					return
				}
				if err := service.ConsumeAward(msg.Body); err != nil {
					if logger != nil {
						logger.Warn("consume activity award failed", zap.Error(err))
					}
					_ = msg.Nack(false, true)
					continue
				}
				_ = msg.Ack(false)
			}
		}
	}()
}

func StartActivitySchedulers(ctx context.Context, service *ActivityLotteryService, logger *zap.Logger) {
	if service == nil {
		return
	}
	go runActivityTicker(ctx, time.Minute, func() {
		if err := service.RetryPendingMessages(ctx, 20, 5); err != nil && logger != nil {
			logger.Warn("activity message retry failed", zap.Error(err))
		}
	})
	go runActivityTicker(ctx, time.Minute, func() {
		if err := service.RefreshPrizePool(ctx); err != nil && logger != nil {
			logger.Warn("activity prize pool refresh failed", zap.Error(err))
		}
	})
}

func runActivityTicker(ctx context.Context, interval time.Duration, fn func()) {
	fn()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}
