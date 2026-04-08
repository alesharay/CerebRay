package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/aray/cerebray/backend/internal/auth"
	"github.com/aray/cerebray/backend/internal/handlers"
	mw "github.com/aray/cerebray/backend/internal/middleware"
)

type RouterDeps struct {
	AuthHandler   *auth.Handler
	Sessions      *auth.SessionStore
	Notes         *handlers.NoteHandlers
	Tags          *handlers.TagHandlers
	Connections   *handlers.ConnectionHandlers
	Dashboard     *handlers.DashboardHandlers
	Glossary      *handlers.GlossaryHandlers
	Conversations *handlers.ConversationHandlers
	Chat          *handlers.ChatHandlers
	AllowOrigin   string
}

func buildRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(mw.CORS(deps.AllowOrigin))

	// Public routes
	r.Get("/health", handlers.HandleHealth)

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", deps.AuthHandler.HandleLogin)
		r.Get("/callback", deps.AuthHandler.HandleCallback)
		r.Post("/logout", deps.AuthHandler.HandleLogout)
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.RequireAuth(deps.Sessions))

		// Auth
		r.Get("/auth/me", deps.AuthHandler.HandleMe)

		// Dashboard
		r.Get("/dashboard", deps.Dashboard.Stats)

		// Notes
		r.Route("/notes", func(r chi.Router) {
			r.Get("/", deps.Notes.List)
			r.Post("/", deps.Notes.Create)
			r.Get("/search", deps.Notes.Search)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", deps.Notes.Get)
				r.Put("/", deps.Notes.Update)
				r.Delete("/", deps.Notes.Delete)
				r.Post("/promote", deps.Notes.Promote)
				r.Post("/sleep", deps.Notes.Sleep)
				r.Post("/archive", deps.Notes.Archive)
				r.Get("/connections", deps.Connections.ListForNote)
				r.Post("/tags", deps.Tags.AddToNote)
				r.Delete("/tags/{tagId}", deps.Tags.RemoveFromNote)
			})
		})

		// Tags
		r.Get("/tags", deps.Tags.List)

		// Connections
		r.Route("/connections", func(r chi.Router) {
			r.Post("/", deps.Connections.Create)
			r.Delete("/{id}", deps.Connections.Delete)
		})

		// Index (graph data)
		r.Get("/index", deps.Connections.GraphData)

		// Conversations
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", deps.Conversations.List)
			r.Post("/", deps.Conversations.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", deps.Conversations.Get)
				r.Delete("/", deps.Conversations.Delete)
				r.Post("/messages", deps.Chat.SendMessage)
			})
		})

		// Settings
		r.Route("/settings", func(r chi.Router) {
			r.Get("/usage", deps.Chat.GetUsage)
		})

		// Glossary
		r.Route("/glossary", func(r chi.Router) {
			r.Get("/", deps.Glossary.List)
			r.Post("/", deps.Glossary.Create)
			r.Put("/{id}", deps.Glossary.Update)
			r.Delete("/{id}", deps.Glossary.Delete)
		})
	})

	return r
}
