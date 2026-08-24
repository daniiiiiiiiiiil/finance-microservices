package http

import (
	"backend/docs"
	"backend/pkg/logger"
	core_http_middleware "backend/pkg/middleware/http"
	"context"
	"errors"
	"fmt"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux        *http.ServeMux
	config     Config
	log        *logger.Logger
	middleware []core_http_middleware.Middleware
}

func NewHTTPServer(config Config, log *logger.Logger, middleware ...core_http_middleware.Middleware) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		config:     config,
		log:        log,
		middleware: middleware,
	}
}

func (s *HTTPServer) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		s.mux.Handle(pattern, route.WithMiddleware())
	}
}

func (s *HTTPServer) RegisterRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.apiVersion)
		s.mux.Handle(prefix+"/", http.StripPrefix(prefix, router.WithMiddleware()))
	}
}

func (s *HTTPServer) RegisterSwagger() {
	s.mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.UIConfig(map[string]string{
			"requestInterceptor": `(req) => { req.credentials = 'include'; return req; }`,
		}),
	))
	s.mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(docs.SwaggerInfo.ReadDoc())); err != nil {
			s.log.Error("failed to write swagger doc", zap.Error(err))
		}
	})
}

func (h *HTTPServer) Run(ctx context.Context) error {
	mux := core_http_middleware.ChainMiddlewares(h.mux, h.middleware...)

	server := &http.Server{
		Addr:    h.config.Address,
		Handler: mux,
	}
	ch := make(chan error, 1)
	go func() {
		defer close(ch)

		h.log.Warn("starting http server", zap.String("address", h.config.Address))

		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		h.log.Warn("Shutdown HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), h.config.ShutDownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		h.log.Warn("HTTP server stopped")
	}
	return nil
}
