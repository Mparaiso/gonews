// Providers are resuable components.
// A provider is defined once and can be reused in any application
// For instance, a SessionProvider might provide session capabilities to an application
// Providers can depend on other providers. The dependencies should be made explicite
// with a provider constructor.
// Providers are meants to be embedded in a single struct called container.
// If a provider depends on other providers then the container is meants to be passed
// as a dependency instead of other providers.

package gonews

import (
	"net/http"

	"bytes"
	"encoding/json"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"os"
)

// FormDecoderProvider provides a FormDecoder to a container
type FormDecoderProvider interface {
	GetFormDecoder() FormDecoder
}

// DefaultFormDecoderProvider provides a FormDecoder to a container
type DefaultFormDecoderProvider struct {
	FormDecoder
}

// NewDefaultFormDecoderProvider returns a new DefaultFormDecoderProvider
func NewDefaultFormDecoderProvider(formDecoder FormDecoder) *DefaultFormDecoderProvider {
	return &DefaultFormDecoderProvider{formDecoder}
}

// GetFormDecoder returns a FormDecoder
func (provider *DefaultFormDecoderProvider) GetFormDecoder() FormDecoder {
	if provider.FormDecoder == nil {
		provider.FormDecoder = NewDefaultFormDecoder()
	}
	return provider.FormDecoder
}

// LoggerProvider provides a logger to a container
type LoggerProvider interface {
	GetLogger() (LoggerInterface, error)
	MustGetLogger() LoggerInterface
}

// DefaultLoggerProvider provides logging capabilies to a container
// it implements LoggerProvider
type DefaultLoggerProvider struct {
	loggerFactory func() (LoggerInterface, error)
	logger        LoggerInterface
	logLevel      LogLevel
	debug         bool
}

// NewDefaultLoggerProvider returns a new DefaultLoggerProvider
func NewDefaultLoggerProvider(loggerFactory func() (LoggerInterface, error), logLevel LogLevel, debug bool) *DefaultLoggerProvider {
	return &DefaultLoggerProvider{loggerFactory: loggerFactory, logLevel: logLevel, debug: debug}
}

// GetLogger gets a logger
func (provider *DefaultLoggerProvider) GetLogger() (LoggerInterface, error) {
	if provider.logger == nil {
		if provider.loggerFactory != nil {
			logger, err := provider.loggerFactory()
			if err != nil {
				return nil, err
			}
			provider.logger = logger
		} else {
			// adhoc logger creation
			logger := &log.Logger{}
			logger.SetOutput(os.Stdout)
			if provider.debug == true {
				provider.logger = NewDefaultLogger(ALL)
			} else {
				provider.logger = NewDefaultLogger(provider.logLevel)
			}

		}
	}
	return provider.logger, nil
}

// MustGetLogger panics on error or return a LoggerInterface
func (provider *DefaultLoggerProvider) MustGetLogger() LoggerInterface {
	logger, err := provider.GetLogger()
	if err != nil {
		panic(err)
	}
	return logger
}

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

// SessionStoreProvider provides a session store
// to a container
type SessionStoreProvider interface {
	GetSessionStore() (sessions.Store, error)
}

// SessionProvider provides a session
// mechanism to a container
type SessionProvider interface {
	GetSession() (SessionWrapper, error)
	MustGetSession() SessionWrapper
}

// DefaultSessionProvider is the default implementation of
// SessionProvider and provides a session
// mechanism to a container
type DefaultSessionProvider struct {
	sessionName               string
	sessionStoreProvider      SessionStoreProvider
	requestProvider           RequestProvider
	responseWithExtraProvider ResponseWriterExtraProvider
	session                   SessionWrapper
}

// NewDefaultSessionProvider returns a *DefaultSessionProvider
func NewDefaultSessionProvider(
	name string, sessionStoreProvider SessionStoreProvider, requestProvider RequestProvider, responseWithExtraProvider ResponseWriterExtraProvider) *DefaultSessionProvider {
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

// GetCSRFGenerator returns the csrf generator
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

// MustGetCSRFGenerator panics on error
func (provider *DefaultCSRFGeneratorProvider) MustGetCSRFGenerator() CSRFGenerator {
	if csrf, err := provider.GetCSRFGenerator(); err != nil {
		panic(err)
	} else {
		return csrf
	}
}

// TemplateProvider provides TemplateEngine
// to a container
type TemplateProvider interface {
	GetTemplate() (TemplateEngine, error)
	MustGetTemplate() TemplateEngine
}

// DefaultTemplateProvider is a default implementation of
// TemplateProvider. It provides templates to a
// container.
type DefaultTemplateProvider struct {
	template TemplateEngine
	templateDirectory,
	templateFileExtension string
	isDebug bool
	LoggerProvider
}

// NewDefaultTemplateProvider creates a new DefaultTemplateProvider
func NewDefaultTemplateProvider(templateDirectory, templateFileExtension string, isDebug bool, loggerProvider LoggerProvider) *DefaultTemplateProvider {
	return &DefaultTemplateProvider{templateDirectory: templateDirectory,
		templateFileExtension: templateFileExtension,
		isDebug:               isDebug,
		LoggerProvider:        loggerProvider,
	}
}

// GetTemplate returns *template.Template
func (provider *DefaultTemplateProvider) GetTemplate() (TemplateEngine, error) {
	if provider.template == nil {
		tpl, err := template.New("templates").Funcs(template.FuncMap{
			"Plus": func(i, j int) int {
				return i + j
			},
			"IsDebug": func() bool {
				return provider.isDebug
			},
			"ToJson": func(object Any) (string, error) {
				b, err := json.MarshalIndent(object, "", "\t")
				if err != nil {
					return "", err
				}
				return bytes.NewBuffer(b).String(), err
			},
		}).ParseGlob(provider.templateDirectory + "/*" + provider.templateFileExtension)

		if err != nil {
			return nil, err
		}
		provider.template = &DefaultTemplateEngine{Template: tpl}
	}
	return provider.template, nil
}

// MustGetTemplate panics on error
func (provider *DefaultTemplateProvider) MustGetTemplate() TemplateEngine {
	tpl, err := provider.GetTemplate()
	if err != nil {
		panic(err)
	}
	return tpl
}
