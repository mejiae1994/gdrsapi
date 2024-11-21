package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gdrsapi/external/gsheets"
	"gdrsapi/internal/gamedocgen"
	"gdrsapi/internal/steamrating"
	"gdrsapi/pkg/config"
	"gdrsapi/pkg/limiter"
	"gdrsapi/pkg/logger"
)

func (app *App) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func (s *App) encodeJsonResponse(w http.ResponseWriter, apiResp *ApiResponse, statusCode int) error {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apiResp); err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return fmt.Errorf("encode json: %w", err)
	}
	s.logger.InfoLog.Println("Sent json encoded response")
	return nil
}

func (app *App) generategdDocument(w http.ResponseWriter, req *http.Request) {
	apiResp := &ApiResponse{
		Result:       nil,
		Sucess:       false,
		ErrorMessage: "",
	}

	// check if method is POST
	if req.Method != http.MethodPost {
		apiResp.ErrorMessage = "Only POST method is allowed"
		err := app.encodeJsonResponse(w, apiResp, http.StatusMethodNotAllowed)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	// now check if we can process the request body
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		apiResp.ErrorMessage = "Form body is too large"
		err := app.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	formData := req.PostForm
	action := formData.Get("action")
	template := formData.Get("template")
	gameTitle := formData.Get("title")
	gameDescription := formData.Get("description")
	gameGenre := formData.Get("genre")
	currentDocumentJsonString := formData.Get("currentDocument")
	// these are optional for regeneration action
	suggestion := formData.Get("suggestion")
	selection := formData.Get("selection")

	// if you have a selection, do you need a suggestion?
	if suggestion != "" && selection != "" {
		app.logger.InfoLog.Println("suggestion and selection provided")
	}

	//validate action and template are present
	if action == "" || template == "" {
		apiResp.ErrorMessage = "both action and template are required"
		err := app.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	validGenFields := gameTitle != "" && gameDescription != "" && gameGenre != ""
	if action == "generate" && !validGenFields {
		apiResp.ErrorMessage = "Required fields are missing"
		err := app.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	if action == "regenerate" && currentDocumentJsonString == "" {
		apiResp.ErrorMessage = "Existing document is required for regeneration"
		err := app.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	var fResp interface{}
	var err error

	switch action {
	case "generate":
		fResp, err = app.documentSvc.GenerateGameDesignDoc(gameTitle, gameDescription, gameGenre, template)
	case "regenerate":
		fResp, err = app.documentSvc.RegenerateGameDesignDoc(currentDocumentJsonString, selection, suggestion, template)
	}

	if err != nil {
		apiResp.ErrorMessage = err.Error()
		err = app.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	apiResp.Result = fResp
	apiResp.Sucess = true
	err = app.encodeJsonResponse(w, apiResp, http.StatusOK)
	if err != nil {
		app.logger.ErrorLog.Println(err.Error())
	}
}

func (s *App) getSteamRating(w http.ResponseWriter, req *http.Request) {
	apiResp := &ApiResponse{
		Result:       nil,
		Sucess:       false,
		ErrorMessage: "",
	}

	// check if method is POST
	if req.Method != http.MethodPost {
		apiResp.ErrorMessage = "Only POST method is allowed"
		err := s.encodeJsonResponse(w, apiResp, http.StatusMethodNotAllowed)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	// now check if we can process the request body
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		apiResp.ErrorMessage = "Form body is too large"
		err := s.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	// get url and validate it
	const steamBaseUrl = "store.steampowered.com"

	steamUrl := req.PostFormValue("url")
	if steamUrl == "" || steamUrl == " " {
		apiResp.ErrorMessage = "Steam Url is required"
		err := s.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	//grab individual meta data from steam url
	steamPageParts := strings.Split(steamUrl, "//")[1]
	baseUrl := strings.Split(steamPageParts, "/")[0]
	gameAppId := strings.Split(steamPageParts, "/")[2]
	gameTitle := strings.Split(steamPageParts, "/")[3]

	if baseUrl != steamBaseUrl {
		apiResp.ErrorMessage = "Steam page Url is invalid"
		err := s.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	//scrape and parse html for steam page content
	steamPgContent, err := s.scrapingSvc.ScrapeSteamPage(steamUrl)
	if err != nil {
		apiResp.ErrorMessage = "Error scraping and parsing steam page"
		err := s.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	se := &gsheets.SheetsEntry{
		Title:      gameTitle,
		AppId:      gameAppId,
		Url:        steamUrl,
		PromptType: "default",
	}

	fResp, err := s.ratingSvc.GetSteamPageRating(*steamPgContent, se)
	if err != nil {
		apiResp.ErrorMessage = err.Error()
		err = s.encodeJsonResponse(w, apiResp, http.StatusBadRequest)
		if err != nil {
			s.logger.ErrorLog.Println(err.Error())
		}
		return
	}

	go s.sheetsSvc.InsertSteamRatingEntry(*se)

	apiResp.Result = fResp
	apiResp.Sucess = true
	err = s.encodeJsonResponse(w, apiResp, http.StatusOK)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
	}
}

type ApiResponse struct {
	Sucess       bool        `json:"success"`
	ErrorMessage string      `json:"errorMessage"`
	Result       interface{} `json:"result"`
}

type App struct {
	scrapingSvc *steamrating.SteamScraper
	ratingSvc   *steamrating.SteamRater
	sheetsSvc   *gsheets.SheetsApp
	documentSvc *gamedocgen.GameDesignDocGen
	logger      *logger.AppLogger
	mu          *sync.Mutex
	limiter     *limiter.Limiter
	cfg         *config.Config
}

func newApp() *App {
	AppLogger := logger.NewAppLogger()

	cfg, err := config.GetConfig()
	if err != nil {
		AppLogger.ErrorLog.Fatal(err.Error())
	}

	limiter := limiter.NewLimiter()
	scrapingSvc := steamrating.NewSteamScraper(AppLogger)
	ratingSvc := steamrating.NewSteamRater(AppLogger)
	gdDocGen := gamedocgen.NewgdDocGen(AppLogger)
	sheetSvc := gsheets.NewSheetsService()

	return &App{
		scrapingSvc: scrapingSvc,
		ratingSvc:   ratingSvc,
		sheetsSvc:   sheetSvc,
		documentSvc: gdDocGen,
		logger:      AppLogger,
		mu:          &sync.Mutex{},
		limiter:     limiter,
		cfg:         cfg,
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := []string{"http://localhost:4321", "https://gamedevreststop.com"}
		origin := r.Header.Get("Origin")
		for _, o := range allowedOrigins {
			if o == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Middlewares
func (s *App) logRequestMidleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.InfoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *App) rateLimitMiddleware(next http.Handler) http.Handler {
	if app.cfg.Environment == "DEV" {
		app.logger.InfoLog.Println("Rate limiting is disabled in dev mode")
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.logger.ErrorLog.Println(err.Error())
		}
		app.mu.Lock()

		if !app.limiter.ClientAllowed(ip) {
			app.mu.Unlock()
			app.logger.InfoLog.Printf("Clisent %s is rate limited", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}

		app.mu.Unlock()
		app.logger.InfoLog.Printf("Client %s is allowed", ip)
		next.ServeHTTP(w, r)
	})
}

func (app *App) checkRateLimitClient() {
	go func() {
		app.logger.InfoLog.Println("Starting remove clients routine")

		for {
			time.Sleep(time.Second * 5)
			app.logger.InfoLog.Println("Removing clients")

			app.mu.Lock()
			app.limiter.RemoveClients()
			app.mu.Unlock()
		}
	}()
}

func (app *App) mapRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/getsteamrating", enableCORS(app.getSteamRating))
	mux.HandleFunc("/gengamedesigndoc", enableCORS(app.generategdDocument))
	mux.HandleFunc("/", app.healthCheck)

	return mux
}

func (app *App) serve() error {
	svc := &http.Server{
		Addr:    ":8082",
		Handler: app.logRequestMidleware(app.rateLimitMiddleware(app.mapRoutes())),
	}

	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.InfoLog.Printf("Received signal %s. Shutting down server", s)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr <- svc.Shutdown(ctx)
		app.logger.InfoLog.Println("Server shutdown successfully")
	}()

	app.logger.InfoLog.Printf("Starting server on port%s", svc.Addr)
	err := svc.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.InfoLog.Println("Stoped server")
	return nil
}

func main() {
	app := newApp()

	err := app.serve()
	if err != nil {
		app.logger.ErrorLog.Fatal(err.Error())
	}
}
