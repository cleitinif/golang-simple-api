package context

import (
	"context"

	"go.uber.org/zap"
)

type ApplicationContext struct {
	context.Context
}

type RequestIDKey struct{}
type LoggerKey struct{}

func NewApplicationContext(ctx context.Context) *ApplicationContext {
	return &ApplicationContext{
		Context: ctx,
	}
}

func (c *ApplicationContext) SetRequestID(requestID string) {
	c.Context = context.WithValue(c.Context, RequestIDKey{}, requestID)
}

func (c *ApplicationContext) GetRequestID() string {
	return c.Value(RequestIDKey{}).(string)
}

func (c *ApplicationContext) SetLogger(logger *zap.SugaredLogger) {
	c.Context = context.WithValue(c.Context, LoggerKey{}, logger)
}

func (c *ApplicationContext) GetLogger() *zap.SugaredLogger {
	return c.Value(LoggerKey{}).(*zap.SugaredLogger)
}
