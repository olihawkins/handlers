/*
Package handlers is a small library of utility http handlers useful for 
building web applications. It includes a NotFoundHandler, an ErrorHandler, 
and a FileHandler. 

The NotFoundHandler and ErrorHandler provide a simple way to respond to the 
client with custom 404 and 500 status pages: create your own application 
specific error page templates and call one of handler's Serve methods with 
the appropriate arguments.

These two handlers are intended to be used indirectly, from inside other 
handlers where server errors or page not found errors occur and you need 
to inform the client. However, both of these types implement the http.Handler 
interface with a ServeHTTP method, which shows their default behaviour when 
bound to a specific route.

The FileHandler provides similar functionality to the FileServer in the 
net/http package, but with two differences: it will not show directory 
listings for directories under its path, and it will respond to any request 
for a non-existent file with the given NotFoundHandler.
*/
package handlers

import (
	"log"
	"os"
	"strings"
	"path/filepath"
	"net/http"
	"html/template"
)

// ErrorMessage holds the message passed to the error template. The template 
// can access the message field with the {{.ErrorMessage}} tag.
type ErrorMessage struct {

	ErrorMessage string
}

// ErrorHandler serves error messages with the given template. The template
// can access the message served by the handler with the {{.ErrorMessage}} tag.
type ErrorHandler struct {

	template *template.Template
	defaultMessage string
	displayErrors bool
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
func NewErrorHandler(t *template.Template, message string, displayErrors bool) *ErrorHandler {

	return &ErrorHandler{

		template: t,
		defaultMessage: message,
		displayErrors: displayErrors,
	}
}

// LoadErrorHandler is a convenience function that returns a new ErrorHandler 
// using the template file specified by tpath. The function first loads the 
// template and then creates the ErrorHandler using NewErrorHandler.
func LoadErrorHandler(tpath string, message string, display bool) *ErrorHandler {

	templateFile, err := template.ParseFiles(tpath)

	if err != nil {
		log.Fatal(err)
	}

	return NewErrorHandler(templateFile, message, display)
}

// ServeError serves the appropriate error message in the error template
// depending on the value of displayErrors. If displayErrors is true then
// the given message is shown, otherwise the default error message is shown.
func (h *ErrorHandler) ServeError(w http.ResponseWriter, message string) {

	var templateData *ErrorMessage
	
	if h.displayErrors {
	
		templateData = &ErrorMessage{message}
	
	} else {

		templateData = &ErrorMessage{h.defaultMessage}
	}

	w.WriteHeader(http.StatusInternalServerError)
	h.template.Execute(w, templateData)
	return
}

// AlwaysServeError serves the given error message in the error template.
// This method overrides the default error message, irrespective of whether 
// displayErrors is false, and ensures that the given message is always shown. 
func (h *ErrorHandler) AlwaysServeError(w http.ResponseWriter, message string) {
	
	templateData := &ErrorMessage{message}
	w.WriteHeader(http.StatusInternalServerError)
	h.template.Execute(w, templateData)
	return
}

// Serve HTTP serves the default error message in the error template.
func (h *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	templateData := &ErrorMessage{h.defaultMessage}
	w.WriteHeader(http.StatusInternalServerError)
	h.template.Execute(w, templateData)
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
func NewNotFoundHandler(t *template.Template) *NotFoundHandler {

	return &NotFoundHandler{

		template: t,
	}
}

// LoadNotFoundHandler is a convenience function that returns a new NotFoundHandler 
// using the template file specified by tpath. The function first loads the 
// template and then creates the NotFoundHandler using NewNotFoundHandler.
func LoadNotFoundHandler(tpath string) *NotFoundHandler {

	templateFile, err := template.ParseFiles(tpath)

	if err != nil {
		log.Fatal(err)
	}

	return NewNotFoundHandler(templateFile)
}

// Serve HTTP serves the path in the handler's template.
func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	templateData := &NotFoundData{r.URL.Path}
	w.WriteHeader(http.StatusNotFound)
	h.template.Execute(w, templateData)
	return
}

// FileHandler serves files requested under the given url path from the given 
// directory. The url path should be the same as the path to which the handler 
// is bound with htp.Handle. If the file is not found the handler serves a 404 
// using the given notFoundHandler. The notFountHandler can be any Handler, but 
// its ServeHTTP method should return a 404. Unlike Go's built-in FileServer, 
// FileHandler will not return directory listings for directories without an 
// index.html and will instead respond with a 404. 
type FileHandler struct {

	urlPath string
	directory string
	notFoundHandler http.Handler
}

// FileHandler returns a new FileHandler with the handler values initialised.
func NewFileHandler(urlPath string, directory string, nfh http.Handler) *FileHandler {

	return &FileHandler{

		urlPath: urlPath,
		directory: directory,
		notFoundHandler: nfh,
	}
}

// Serve HTTP serves the path in the handler's template.
func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	const indexPage string = "index.html"
	
	var(
		requestPath string = r.URL.Path[len(h.urlPath)-1:] 
		filePath string
	)
	
	// If the request path ends in "/" ...
	if strings.HasSuffix(r.URL.Path, "/") {

		// Set the target filepath to index.html
		filePath = h.directory + filepath.FromSlash(requestPath + indexPage)	
	
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

	// Check the mode to ensure the target file path is a file
	switch mode := finfo.Mode(); {

	// If the target file is a directory redirect to the path with a slash
	case mode.IsDir():
		
		http.Redirect(w, r, r.URL.Path + "/", http.StatusFound)
	
	// Otherwise serve the file
	case mode.IsRegular():
		
		http.ServeFile(w, r, filePath)
	}

	return
}
	