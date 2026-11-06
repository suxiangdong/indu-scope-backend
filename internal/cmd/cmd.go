package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminAttributesController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminattributes"
	adminAuthController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminauth"
	adminCategoriesController "github.com/suxiangdong/indu-scope-backend/internal/controller/admincategories"
	adminCustomerUsersController "github.com/suxiangdong/indu-scope-backend/internal/controller/admincustomerusers"
	adminStorePlatformController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminstoreplatform"
	adminUsersController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminusers"
	adminVerificationController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminverification"
	"github.com/suxiangdong/indu-scope-backend/internal/controller/auth"
	supplierProductsController "github.com/suxiangdong/indu-scope-backend/internal/controller/supplierproducts"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
	"github.com/suxiangdong/indu-scope-backend/internal/sessionstore"
	"github.com/suxiangdong/indu-scope-backend/internal/wechat"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			db := g.DB()
			dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := db.Ctx(dbCtx).PingMaster(); err != nil {
				_ = db.Close(context.Background())
				return gerror.Wrap(err, "ping database")
			}

			redis := g.Redis()
			redisCtx, cancelRedis := context.WithTimeout(ctx, 5*time.Second)
			defer cancelRedis()
			if _, err := redis.Do(redisCtx, "PING"); err != nil {
				_ = redis.Close(context.Background())
				_ = db.Close(context.Background())
				return gerror.Wrap(err, "ping redis")
			}
			if err := wechat.WarmupAccessToken(ctx); err != nil {
				// Keep the API available when WeChat is temporarily unreachable;
				// getAccessToken will retry lazily when a phone login is requested.
				g.Log().Warningf(ctx, "wechat access token warmup failed: %v", err)
			}
			closeResources := func() error {
				var closeErr error
				if err := redis.Close(context.Background()); err != nil {
					closeErr = gerror.Wrap(err, "close redis")
				}
				if err := db.Close(context.Background()); err != nil && closeErr == nil {
					closeErr = gerror.Wrap(err, "close database")
				}
				return closeErr
			}

			s := g.Server()
			s.SetSessionMaxAge(0)
			s.SetSessionCookieOutput(false)
			s.SetSessionStorage(sessionstore.NewNoExpireRedisStorage(redis, consts.RedisSessionPrefix))
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.Ctx)
				group.Bind(auth.NewV1())
				group.Bind(supplierProductsController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.AdminCtx)
				group.Bind(adminAuthController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.AdminCtx, middleware.RequireAdmin)
				group.Bind(adminUsersController.NewV1())
				group.Bind(adminCustomerUsersController.NewV1())
				group.Bind(adminVerificationController.NewV1())
				group.Bind(adminCategoriesController.NewV1())
				group.Bind(adminAttributesController.NewV1())
				group.Bind(adminStorePlatformController.NewV1())
			})

			shutdownCtx, stopSignal := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stopSignal()
			if err := s.Start(); err != nil {
				_ = closeResources()
				return err
			}
			<-shutdownCtx.Done()

			shutdownErr := s.Shutdown()
			closeErr := closeResources()
			if shutdownErr != nil {
				return gerror.Wrap(shutdownErr, "shutdown http server")
			}
			if closeErr != nil {
				return closeErr
			}
			return nil
		},
	}
)
