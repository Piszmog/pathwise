package mux

import (
	"net/http"
)

func NewMux(options ...Option) *http.ServeMux {
	mux := http.NewServeMux()
	for _, o := range options {
		o(mux)
	}
	return mux
}

type Option func(*http.ServeMux)

func WithHandle(method string, path string, handler http.Handler) Option {
	return func(s *http.ServeMux) {
		s.Handle(method+" "+path, handler)
	}
}

func WithGeneralHandle(path string, handler http.Handler) Option {
	return func(s *http.ServeMux) {
		s.Handle(path, handler)
	}
}

func WithHandleFunc(method string, path string, handleFunc http.HandlerFunc) Option {
	return func(s *http.ServeMux) {
		s.HandleFunc(method+" "+path, handleFunc)
	}
}
