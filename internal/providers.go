package gonews

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// ResponseWriterExtraProvider provides a ResponseWriterExtra
type ResponseWriterExtraProvider interface {
	ResponseWriter() ResponseWriterExtra
}

// RequestProvider provides an *http.Request
type RequestProvider interface {
	Request() *http.Request
}

// SecretProvider provides a secret key
type SecretProvider interface {
	GetSecret() string
}

type SessionStoreProvider interface {
	GetSessionStore() (sessions.Store, error)
}

type SessionProvider interface {
	GetSession() (SessionWrapper, error)
	MustGetSession() SessionWrapper
}

type DefaultSessionProvider struct {
	sessionName               string
	sessionStoreProvider      SessionStoreProvider
	requestProvider           RequestProvider
	responseWithExtraProvider ResponseWriterExtraProvider
	session                   SessionWrapper
}

// NewDefaultSessionProvider returns a *DefaultSessionProvider
func NewDefaultSessionProvider(
	name string,
	sessionStoreProvider SessionStoreProvider,
	requestProvider RequestProvider,
	responseWithExtraProvider ResponseWriterExtraProvider,
) *DefaultSessionProvider {
	return &DefaultSessionProvider{name, sessionStoreProvider, requestProvider, responseWithExtraProvider, nil}
}

// GetSession returns the session
func (provider *DefaultSessionProvider) GetSession() (SessionWrapper, error) {
	if provider.session == nil {
		sessionStore, err := provider.sessionStoreProvider.GetSessionStore()
		if err != nil {
			return nil, err
		}
		session, err := NewSession(sessionStore, provider.requestProvider.Request(), provider.sessionName)
		if err != nil {
			return nil, err
		}

		provider.session = session
		provider.session.SetOptions(&sessions.Options{
			Path:     "/",
			HttpOnly: true,
			// Secure:   true,
			MaxAge: 60 * 60 * 24,
			Domain: provider.requestProvider.Request().URL.Host,
		})
		provider.responseWithExtraProvider.ResponseWriter().SetSession(provider.session)
	}
	return provider.session, nil
}

// MustGetSession panics on error
func (provider *DefaultSessionProvider) MustGetSession() SessionWrapper {
	session, err := provider.GetSession()
	if err != nil {
		panic(err)
	}
	return session
}

// CSRFGeneratorProvider provides a CSRF generator
type CSRFGeneratorProvider interface {
	// GetCSRF returns the csrf
	GetCSRFGenerator() (CSRFGenerator, error)
	// MustGetCSRF panics on error
	MustGetCSRFGenerator() CSRFGenerator
}

// DefaultCSRFGeneratorProvider implements CSRFGeneratorProvider
type DefaultCSRFGeneratorProvider struct {
	sessionProvider SessionProvider
	secretProvider  SecretProvider
	csrfGenerator   CSRFGenerator
}

// NewDefaultCSRFGeneratorProvider returns a new DefaultCSRFGeneratorProvider
func NewDefaultCSRFGeneratorProvider(sessionProvider SessionProvider, secretProvider SecretProvider) *DefaultCSRFGeneratorProvider {
	return &DefaultCSRFGeneratorProvider{sessionProvider, secretProvider, nil}
}

// GetCSRF returns the csrf
func (provider *DefaultCSRFGeneratorProvider) GetCSRFGenerator() (CSRFGenerator, error) {
	if provider.csrfGenerator == nil {
		session, err := provider.sessionProvider.GetSession()
		if err != nil {
			return nil, err
		}
		provider.csrfGenerator = &DefaultCSRFGenerator{session, provider.secretProvider.GetSecret()}
	}
	return provider.csrfGenerator, nil
}

// MustGetCSRF panics on error
func (provider *DefaultCSRFGeneratorProvider) MustGetCSRFGenerator() CSRFGenerator {
	if csrf, err := provider.GetCSRFGenerator(); err != nil {
		panic(err)
	} else {
		return csrf
	}
}
