package server

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	authhnd "github.com/nathakusuma/auditorium-reservation-backend/internal/app/auth/handler"
	authrepo "github.com/nathakusuma/auditorium-reservation-backend/internal/app/auth/repository"
	authsvc "github.com/nathakusuma/auditorium-reservation-backend/internal/app/auth/service"
	conferencehnd "github.com/nathakusuma/auditorium-reservation-backend/internal/app/conference/handler"
	conferencerepo "github.com/nathakusuma/auditorium-reservation-backend/internal/app/conference/repository"
	conferencesvc "github.com/nathakusuma/auditorium-reservation-backend/internal/app/conference/service"
	feedbackhnd "github.com/nathakusuma/auditorium-reservation-backend/internal/app/feedback/handler"
	feedbackrepo "github.com/nathakusuma/auditorium-reservation-backend/internal/app/feedback/repository"
	feedbacksvc "github.com/nathakusuma/auditorium-reservation-backend/internal/app/feedback/service"
	registrationhnd "github.com/nathakusuma/auditorium-reservation-backend/internal/app/registration/handler"
	registrationrepo "github.com/nathakusuma/auditorium-reservation-backend/internal/app/registration/repository"
	registrationsvc "github.com/nathakusuma/auditorium-reservation-backend/internal/app/registration/service"
	userhnd "github.com/nathakusuma/auditorium-reservation-backend/internal/app/user/handler"
	userrepo "github.com/nathakusuma/auditorium-reservation-backend/internal/app/user/repository"
	usersvc "github.com/nathakusuma/auditorium-reservation-backend/internal/app/user/service"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/infra/env"
	"github.com/nathakusuma/auditorium-reservation-backend/internal/middleware"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/jwt"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/mail"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/supabase"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/uuidpkg"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/validator"
	"github.com/redis/go-redis/v9"
)

type HttpServer interface {
	Start(part string)
	MountMiddlewares()
	MountRoutes(db *sqlx.DB, rds *redis.Client)
	GetApp() *fiber.App
}

type httpServer struct {
	app *fiber.App
}

func NewHttpServer() HttpServer {
	config := fiber.Config{
		AppName:      "Auditorium Reservation",
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ErrorHandler: ErrorHandler(),
	}

	app := fiber.New(config)

	return &httpServer{
		app: app,
	}
}

func (s *httpServer) GetApp() *fiber.App {
	return s.app
}

func (s *httpServer) Start(port string) {
	if port[0] != ':' {
		port = ":" + port
	}

	err := s.app.Listen(port)

	if err != nil {
		log.Fatal(map[string]interface{}{
			"error": err.Error(),
		}, "[SERVER][Start] failed to start server")
	}
}

func (s *httpServer) MountMiddlewares() {
	// s.app.Use(middleware.LoggerConfig())
	s.app.Use(middleware.Helmet())
	s.app.Use(middleware.Compress())
	s.app.Use(middleware.Cors())
	s.app.Use(middleware.RecoverConfig())
}

func (s *httpServer) MountRoutes(db *sqlx.DB, rds *redis.Client) {
	// Deleted for Cryptographic Failures.
	// bcryptInstance := bcrypt.GetBcrypt()
	jwtAccess := jwt.NewJwt(env.GetEnv().JwtAccessExpireDuration, env.GetEnv().JwtAccessSecretKey)
	mailer := mail.NewMailDialer()
	uuidInstance := uuidpkg.GetUUID()
	validatorInstance := validator.NewValidator()
	middlewareInstance := middleware.NewMiddleware(jwtAccess)
	supabase := supabase.New()

	s.app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString("Healthy")
	})

	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	userRepository := userrepo.NewUserRepository(db)
	authRepository := authrepo.NewAuthRepository(db, rds)
	conferenceRepository := conferencerepo.NewConferenceRepository(db)
	registrationRepository := registrationrepo.NewRegistrationRepository(db)
	feedbackRepository := feedbackrepo.NewFeedbackRepository(db)

	userService := usersvc.NewUserService(userRepository, supabase, uuidInstance)
	authService := authsvc.NewAuthService(authRepository, userService, jwtAccess, mailer, uuidInstance)
	conferenceService := conferencesvc.NewConferenceService(conferenceRepository, uuidInstance)
	registrationService := registrationsvc.NewRegistrationService(registrationRepository, conferenceService)
	feedbackService := feedbacksvc.NewFeedbackService(feedbackRepository, registrationService, conferenceService,
		uuidInstance)

	userhnd.InitUserHandler(v1, middlewareInstance, validatorInstance, userService)
	authhnd.InitAuthHandler(v1, middlewareInstance, validatorInstance, authService)
	conferencehnd.InitConferenceHandler(v1, middlewareInstance, validatorInstance, conferenceService)
	registrationhnd.InitRegistrationHandler(v1, middlewareInstance, validatorInstance, registrationService)
	feedbackhnd.InitFeedbackHandler(v1, middlewareInstance, validatorInstance, feedbackService)
}
