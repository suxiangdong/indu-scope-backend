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

	"github.com/suxiangdong/indu-scope-backend/internal/authutil"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminAnalyticsController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminanalytics"
	adminAppContentController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminappcontent"
	adminAttributesController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminattributes"
	adminAuthController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminauth"
	adminCategoriesController "github.com/suxiangdong/indu-scope-backend/internal/controller/admincategories"
	adminCustomerUsersController "github.com/suxiangdong/indu-scope-backend/internal/controller/admincustomerusers"
	adminHotCategoriesController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminhotcategories"
	adminProcurementRequirementsController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminprocurementrequirements"
	adminProductsController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminproducts"
	adminStockUnitController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminstockunit"
	adminStorePlatformController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminstoreplatform"
	adminUsersController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminusers"
	adminVerificationController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminverification"
	adminVipController "github.com/suxiangdong/indu-scope-backend/internal/controller/adminvip"
	analyticsController "github.com/suxiangdong/indu-scope-backend/internal/controller/analytics"
	"github.com/suxiangdong/indu-scope-backend/internal/controller/auth"
	categoriesController "github.com/suxiangdong/indu-scope-backend/internal/controller/categories"
	favoritesController "github.com/suxiangdong/indu-scope-backend/internal/controller/favorites"
	paymentController "github.com/suxiangdong/indu-scope-backend/internal/controller/payment"
	procurementRequirementsController "github.com/suxiangdong/indu-scope-backend/internal/controller/procurementrequirements"
	productCatalogController "github.com/suxiangdong/indu-scope-backend/internal/controller/productcatalog"
	supplierProductsController "github.com/suxiangdong/indu-scope-backend/internal/controller/supplierproducts"
	userProfileController "github.com/suxiangdong/indu-scope-backend/internal/controller/userprofile"
	vipController "github.com/suxiangdong/indu-scope-backend/internal/controller/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
	"github.com/suxiangdong/indu-scope-backend/internal/sessionstore"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
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
			if err := storage.InitOSSClient(ctx); err != nil {
				_ = redis.Close(context.Background())
				_ = db.Close(context.Background())
				return gerror.Wrap(err, "initialize oss client")
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
			configureAPIDocs(ctx, s)
			s.SetSessionMaxAge(authutil.TokenTTL)
			s.SetSessionCookieOutput(false)
			s.SetSessionStorage(sessionstore.NewRedisStorage(redis, consts.RedisSessionPrefix))
			frontendAuthController := auth.NewV1()
			adminAuth := adminAuthController.NewV1()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.POST("/api/v1/payment/wechat/notify", paymentController.Notify)
				group.POST("/api/v1/payment/wechat/refund/notify", paymentController.RefundNotify)
			})

			// Public authentication endpoints. Every other frontend endpoint is
			// registered in one of the protected groups below.
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.RateLimit(time.Minute, 30), middleware.Ctx)
				group.POST("/", frontendAuthController, "Register")
				group.POST("/", frontendAuthController, "Login")
				group.POST("/", frontendAuthController, "WechatPhoneLogin")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.RateLimit(time.Minute, 30), middleware.Ctx)
				group.Bind(categoriesController.NewV1())
			})
			registerPublicAppContentRoutes(s)

			// Verification and account endpoints are available to any logged-in
			// user, including users whose profile status is still 0 (or has been
			// rejected and needs to be submitted again).
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.RateLimit(time.Minute, 30), middleware.Ctx, middleware.RequireAuthenticated)
				group.GET("/", frontendAuthController, "ListStorePlatforms")
				group.GET("/", frontendAuthController, "GetVerification")
				group.POST("/", frontendAuthController, "UploadBusinessLicense")
				group.POST("/", frontendAuthController, "RecognizeBusinessLicense")
				group.POST("/", frontendAuthController, "UploadIDCard")
				group.POST("/", frontendAuthController, "SubmitPersonalVerification")
				group.POST("/", frontendAuthController, "SubmitCompanyVerification")
				group.GET("/", frontendAuthController, "Me")
				group.POST("/", frontendAuthController, "Logout")
			})

			// Supplier product APIs require all three independent conditions:
			// logged in, verified, and supplier role.
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
					middleware.RequireRole(middleware.UserRoleSupplier),
				)
				group.Bind(supplierProductsController.NewV1())
			})

			// Favorites are buyer-only. The role middleware is intentionally
			// separate from RequireVerified so the same building blocks can be
			// reused for future buyer APIs.
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
					middleware.RequireRole(middleware.UserRoleBuyer),
				)
				group.Bind(favoritesController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
				)
				group.Bind(userProfileController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
					middleware.RequireRole(middleware.UserRoleBuyer),
				)
				group.Bind(productCatalogController.NewV1())
			})
			// Procurement requirement reads and management actions are available to
			// verified buyers and suppliers. Publishing a requirement, uploading a
			// requirement attachment, and responding to a requirement additionally
			// require an active VIP subscription.
			procurementRequirements := procurementRequirementsController.NewV1()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
				)
				group.GET("/", procurementRequirements, "FormOptions")
				group.GET("/", procurementRequirements, "List")
				group.GET("/", procurementRequirements, "Detail")
				group.POST("/", procurementRequirements, "UpdateStatus")
				group.POST("/", procurementRequirements, "UpdateResponseStatus")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.Ctx,
					middleware.RequireAuthenticated,
					middleware.RequireVerified,
					middleware.RequireActiveVip,
				)
				group.POST("/", procurementRequirements, "Create")
				group.POST("/", procurementRequirements, "UploadFile")
				group.POST("/", procurementRequirements, "CreateResponse")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.Ctx, middleware.RequireAuthenticated, middleware.RequireVerified)
				group.Bind(analyticsController.NewV1())
				group.Bind(vipController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.RateLimit(time.Minute, 10), middleware.AdminCtx)
				group.POST("/", adminAuth, "Login")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.AdminCtx, middleware.RequireAdmin)
				group.GET("/", adminAuth, "Me")
				group.POST("/", adminAuth, "Logout")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.AdminCtx, middleware.RequireAdmin)
				group.Bind(adminCustomerUsersController.NewV1())
				group.Bind(adminAppContentController.NewV1())
				group.Bind(adminAnalyticsController.NewV1())
				group.Bind(adminVipController.NewV1())
				group.Bind(adminProductsController.NewV1())
				group.Bind(adminProcurementRequirementsController.NewV1())
				group.Bind(adminVerificationController.NewV1())
				group.Bind(adminCategoriesController.NewV1())
				group.Bind(adminHotCategoriesController.NewV1())
				group.Bind(adminAttributesController.NewV1())
				group.Bind(adminStorePlatformController.NewV1())
				group.Bind(adminStockUnitController.NewV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.AdminCtx, middleware.RequireAdmin, middleware.RequireSuperAdmin)
				group.Bind(adminUsersController.NewV1())
				group.Bind(adminVipController.NewRefundV1())
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					middleware.HTTPErrorStatus,
					middleware.AdminCtx,
					middleware.RequireAdmin,
					middleware.RequireSuperAdmin,
				)
				group.POST("/", adminCustomerUsersController.NewRoleV1(), "UpdateRole")
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
