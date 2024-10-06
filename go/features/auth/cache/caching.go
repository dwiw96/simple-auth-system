package chache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	auth "github.com/dwiw96/simple-auth-system/features/auth"
)

type authCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewAuthCache(client *redis.Client, ctx context.Context) auth.CacheInterface {
	return &authCache{
		client: client,
		ctx:    ctx,
	}
}

func (c *authCache) CachingBlockedToken(payload auth.JwtPayload) error {
	iat := time.Unix(payload.Iat, 0)
	exp := time.Unix(payload.Exp, 0)
	duration := time.Duration(exp.Sub(iat).Nanoseconds())
	return c.client.Set(c.ctx, fmt.Sprint("block ", payload.ID), payload.UserID, duration).Err()
}
