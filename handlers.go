/*
Package handlers is a small library of utility http handlers that are useful
for building web applications. It includes a NotFoundHandler, an ErrorHandler,
and a FileHandler.

The NotFoundHandler and ErrorHandler provide a simple way to respond to the
client with custom 404 and 500 status pages: create your own application
specific error page templates and call each of the handler's Serve methods
with the appropriate arguments.

These two handlers are intended to be used indirectly, from inside other
handlers where server errors or page not found errors occur and you need
to report them to the browser. However, both of these types implement the
http.Handler interface with a ServeHTTP method, which shows their default
behaviour when bound to a specific route.

The FileHandler provides similar functionality to the FileServer in the
net/http package, but with two differences: it will not show directory
listings for directories under its path, and it will respond to any request
for a non-existent file with the given NotFoundHandler.
*/
package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ErrorMessage holds the message passed to the error template. The template
// can access the message field with the {{.ErrorMessage}} tag.
type ErrorMessage struct {
	ErrorMessage string
}

// ErrorHandler serves error messages with the given template. The template
// can access the message served by the handler with the {{.ErrorMessage}} tag.
type ErrorHandler struct {
	template       *template.Template
	defaultMessage string
	displayErrors  bool
}

// NewErrorHandler returns a new ErrorHandler with the handler values initialised.
// The handler uses the given template to print an error message. The template
// must display {{.ErrorMessage}}. The default error message is set to message.
// The display argument controls whether error messages passed to the handler's
// ServeError function are shown to the user on the error page, or whether the
// default error message is shown instead. This allows detailed error messages
// to be printed to the screen during development, but turned off in production.
// The handler's AlwaysServeError method forces the display of a particular
// error message even if displayErrors is set to false.
func NewErrorHandler(template *template.Template, defaultMessage string, displayErrors bool) *ErrorHandler {

	return &ErrorHandler{
		template:       template,
		defaultMessage: defaultMessage,
		displayErrors:  displayErrors,
	}
}

// LoadErrorHandler is a convenience function that returns a new ErrorHandler
// using the template file specified by templatePath. The function first loads
// the template and then creates the ErrorHandler using NewErrorHandler.
func LoadErrorHandler(templatePath string, defaultMessage string, displayErrors bool) *ErrorHandler {

	template, err := template.ParseFiles(templatePath)

	if err != nil {
		log.Fatal(err)
	}

	return NewErrorHandler(template, defaultMessage, displayErrors)
}

// ServeError serves the appropriate error message in the error template
// depending on the value of displayErrors. If displayErrors is true then
// the given message is shown, otherwise the default error message is shown.
func (h *ErrorHandler) ServeError(w http.ResponseWriter, message string) {

	var (
		templateData *ErrorMessage
		buffer       bytes.Buffer
	)

	if h.displayErrors {

		templateData = &ErrorMessage{message}

	} else {

		templateData = &ErrorMessage{h.defaultMessage}
	}

	// Execute template into buffer
	err := h.template.Execute(&buffer, templateData)

	// If template execution fails, fall back to the built-in http error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Otherwise serve the error in the error template
	w.WriteHeader(http.StatusInternalServerError)
	buffer.WriteTo(w)
	return
}

// AlwaysServeError serves the given error message in the error template.
// This method overrides the default error message, irrespective of whether
// displayErrors is false, and ensures that the given message is always shown.
func (h *ErrorHandler) AlwaysServeError(w http.ResponseWriter, message string) {

	var buffer bytes.Buffer
	templateData := &ErrorMessage{message}

	// Execute template into buffer
	err := h.template.Execute(&buffer, templateData)

	// If template execution fails, fall back to the built-in http error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Otherwise serve the error in the error template
	w.WriteHeader(http.StatusInternalServerError)
	buffer.WriteTo(w)
	return
}

// ServeHTTP serves the default error message in the error template.
func (h *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var buffer bytes.Buffer
	templateData := &ErrorMessage{h.defaultMessage}

	// Execute template into buffer
	err := h.template.Execute(&buffer, templateData)

	// If template execution fails, fall back to the built-in http error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Otherwise serve the error in the error template
	w.WriteHeader(http.StatusInternalServerError)
	buffer.WriteTo(w)
	return
}

// NotFoundData holds the path passed to the handler's template. The template
// can access the message field with the {{.Path}} tag.
type NotFoundData struct {
	Path string
}

// NotFoundHandler serves a 404 with the given template. The template
// can access the path to the file not found with {{.Path}} tag.
type NotFoundHandler struct {
	template *template.Template
}

// NewNotFoundHandler returns a new NotFoundHandler with the handler values
// initialised. The handler uses the given template to print the path to the
// file not found with a 404. The template must display {{.Path}}.
func NewNotFoundHandler(template *template.Template) *NotFoundHandler {

	return &NotFoundHandler{
		template: template,
	}
}

// LoadNotFoundHandler is a convenience function that returns a new NotFoundHandler
// using the template file specified by templatePath. The function first loads the
// template and then creates the NotFoundHandler using NewNotFoundHandler.
func LoadNotFoundHandler(templatePath string) *NotFoundHandler {

	template, err := template.ParseFiles(templatePath)

	if err != nil {
		log.Fatal(err)
	}

	return NewNotFoundHandler(template)
}

// ServeHTTP serves the path in the handler's template.
func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var buffer bytes.Buffer
	templateData := &NotFoundData{r.URL.Path}

	// Execute template into buffer
	err := h.template.Execute(&buffer, templateData)

	// If template execution fails, report it with the built-in http error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Otherwise serve the 404 in the not found template
	w.WriteHeader(http.StatusNotFound)
	buffer.WriteTo(w)
	return
}

// FileHandler serves files requested under the given url path from the given
// directory. The url path should be the same as the path to which the handler
// is bound with http.Handle. If the file is not found the handler serves a 404
// using the given notFoundHandler. The notFoundHandler can be any Handler, but
// its ServeHTTP method should return a 404. Unlike Go's built-in FileServer,
// FileHandler will not return directory listings for directories without an
// index.html and will instead respond with a 404.
type FileHandler struct {
	urlPath         string
	directory       string
	notFoundHandler http.Handler
}

// NewFileHandler returns a new FileHandler with the handler values initialised.
func NewFileHandler(urlPath string, directory string, notFoundHandler http.Handler) *FileHandler {

	return &FileHandler{
		urlPath:         urlPath,
		directory:       directory,
		notFoundHandler: notFoundHandler,
	}
}

// ServeHTTP is a wrapper around http.ServeFile, with paths and response
// values modified to provide the appropriate behaviour for the FileHandler.
func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	const indexPage string = "index.html"

	var (
		requestPath string = r.URL.Path[len(h.urlPath)-1:]
		filePath    string
	)

	// If the request path ends in "/" ...
	if strings.HasSuffix(r.URL.Path, "/") {

		// Set the target filepath to index.html
		filePath = h.directory + filepath.FromSlash(requestPath+indexPage)

	} else {

		// Otherwise set the target filepath to the named file
		filePath = h.directory + filepath.FromSlash(requestPath)
	}

	// Try to get file info
	finfo, err := os.Stat(filePath)

	// If Stat fails return a 404
	if err != nil {

		h.notFoundHandler.ServeHTTP(w, r)
		return
	}

	// Check the mode to ensure the target filepath is a file
	switch mode := finfo.Mode(); {

	// If the target file is a directory redirect to the path with a slash
	case mode.IsDir():

		http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)

	// Otherwise serve the file
	case mode.IsRegular():

		http.ServeFile(w, r, filePath)
	}

	return
}
