package env

import (
	"context"
	"errors"
)

type envKeyType int

const envKey envKeyType = 1

// NewContext creates a new context with the given env
func NewContext(ctx context.Context, env Env) context.Context {
	return context.WithValue(ctx, envKey, env)
}

// FromContext returns the environment from a context
func FromContext(ctx context.Context) (Env, error) {
	if env, ok := ctx.Value(envKey).(Env); ok {
		return env, nil
	}
	return nil, errors.New("env not found")
}
