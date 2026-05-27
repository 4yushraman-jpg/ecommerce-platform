package service

import (
	"context"

	"github.com/4yushraman-jpg/product-service/internal/model"
)

type EventPublisher interface {
	Publish(ctx context.Context, event model.ProductEvent) error
}

type NoopEventPublisher struct{}

func (p *NoopEventPublisher) Publish(ctx context.Context, event model.ProductEvent) error {
	return nil
}
