package factory

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authCache "github.com/dwiw96/simple-auth-system/features/auth/cache"
	authDelivery "github.com/dwiw96/simple-auth-system/features/auth/delivery"
	authRepository "github.com/dwiw96/simple-auth-system/features/auth/repository"
	authService "github.com/dwiw96/simple-auth-system/features/auth/service"
)

func InitFactory(router *gin.Engine, pool *pgxpool.Pool, rdClient *redis.Client, ctx context.Context) {
	authRepoInterface := authRepository.NewAuthRepository(pool, ctx)
	authCacheInterface := authCache.NewAuthCache(rdClient, ctx)
	authServiceInterface := authService.NewAuthService(authRepoInterface, authCacheInterface)
	authDelivery.NewAuthDelivery(router, authServiceInterface, pool, rdClient, ctx)
}
