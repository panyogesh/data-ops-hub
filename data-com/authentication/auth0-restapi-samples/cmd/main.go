package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/auth0-restapi-samples/auth"
	basicauth "github.com/auth0-restapi-samples/basics"
	"github.com/auth0-restapi-samples/config"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"
)

// Helper to replace gin.H / c.JSON
func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RestApiInitializer(chr *chi.Mux, config config.Config) {
	chr.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./config/openapi-authentication.yaml")
	})

	// Swagger UI at /docs (points to /openapi.yaml)
	ui := v5emb.New("MyAuthentication API", "/openapi.yaml", "/docs/")
	chr.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		ui.ServeHTTP(w, r)
	})

	authServices := auth.NewAuth(config)

	chr.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.RawQuery
		target := "/docs/"
		if q != "" {
			target = "/docs/?" + q
		}
		http.Redirect(w, r, target, http.StatusFound)
	})

	chr.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		// Define your parameters
		finalURL := authServices.LoginServices()

		// Perform the redirect
		http.Redirect(w, r, finalURL, http.StatusFound)
	})

	chr.Get("/token", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			renderJSON(w, http.StatusBadRequest, map[string]string{"error": "missing query param: code"})
			return
		}

		status, result := authServices.TokenValidation(code)
		renderJSON(w, status, result)
	})

}

func Server(config config.Config) {
	chr := chi.NewRouter()

	chr.Use(middleware.Logger)
	chr.Use(middleware.Recoverer)

	RestApiInitializer(chr, config)

	http.ListenAndServe(":"+strconv.Itoa(config.Server.Port), chr)
}

func main() {
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	basicauth.BasicAuthExample(*config)
	Server(*config)
}
