package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type ctxKey int

const (
	ctxUserKey ctxKey = iota
)

// RootHandler returns a serve mux that handles all routes.
func RootHandler(db *DB) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", &ensureAuthentication{
		handler: apiHandler(db),
		db:      db,
	}))
	return mux
}

type authenticatedHandler func(http.ResponseWriter, *http.Request, *User)
type ensureAuthentication struct {
	handler http.Handler
	db      *DB
}

func (a *ensureAuthentication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var token string
	if r.Header.Get("Authorization") != "" {
		token = r.Header.Get("Authorization")
	} else {
		token = r.URL.Query().Get("token")
	}

	user, err := a.db.userFromToken(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx = context.WithValue(ctx, ctxUserKey, user)
	r = r.WithContext(ctx)

	a.handler.ServeHTTP(w, r)
}

type customHandler func(http.ResponseWriter, *http.Request, *DB, *User)

func wrapHandler(h customHandler, db *DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(ctxUserKey).(*User)
		h(w, r, db, user)
	})
}

// apiV1Handler returns a serve mux that handles API routes.
func apiHandler(db *DB) *ensureAuthentication {
	mux := http.ServeMux{}
	mux.Handle("/eat", wrapHandler(eatHandler, db))
	mux.Handle("/poop", wrapHandler(poopHandler, db))
	return &ensureAuthentication{
		handler: &mux,
		db:      db,
	}
}

type eatRequest struct {
	AteAt time.Time `json:"ate_at"`
}

// eatHandler returns a handler that handles the /eat route in the /api/v1
// namespace. It implements customHandler.
func eatHandler(w http.ResponseWriter, r *http.Request, db *DB, user *User) {
	//ctx := r.Context()

	//var e eatRequest
	if _, err := w.Write([]byte("yum")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type poopRequest struct {
	PoopedAt time.Time `json:"pooped_at"`
	Bad      bool      `json:"quality"`
}

// poopHandler returns a handler that handles the /poop route in the /api/v1
// namespace. It implements customHandler.
func poopHandler(w http.ResponseWriter, r *http.Request, db *DB, user *User) {
	ctx := r.Context()

	var p poopRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.logPoop(ctx, user.ID, p.PoopedAt, p.Bad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
