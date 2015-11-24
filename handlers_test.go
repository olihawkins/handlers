package handlers

import(
	"testing"
	"path/filepath"
	"net/http"
	"net/http/httptest"
)

// Test ErrorHandler functions and methods
func TestErrorHandler(t *testing.T) {

	const(
		defaultMessage string = "Default error message"
		customMessage string = "Test ServeError"
	)

	var(
		h *ErrorHandler
		templatePath string
		bodyString string
		response *httptest.ResponseRecorder
	)	

	// Set the template path
	templatePath = filepath.FromSlash("templates/error.html")

	// Get an ErrorHandler with the example template and display errors on
	h = LoadErrorHandler(templatePath, defaultMessage, true)

	// Test ServeError with display errors on
	response = httptest.NewRecorder()
	h.ServeError(response, customMessage)

	// Check status code
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusInternalServerError from ErrorHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the custom message
	bodyString = response.Body.String()

	if bodyString != "Error: " + customMessage {
		t.Errorf("Expected \"Error: " + customMessage + 
			"\" from ErrorHandler. Got: %s", bodyString)
	}

	// Test AlwaysServeError with display errors on
	response = httptest.NewRecorder()
	h.AlwaysServeError(response, customMessage)

	// Check status code
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusInternalServerError from ErrorHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the custom message
	bodyString = response.Body.String()

	if bodyString != "Error: " + customMessage {
		t.Errorf("Expected \"Error: " + customMessage + 
			"\" from ErrorHandler. Got: %s", bodyString)
	}

	// Get an ErrorHandler with the example template and display errors off
	h = LoadErrorHandler(templatePath, defaultMessage, false)

	// Test ServeError with display errors off
	response = httptest.NewRecorder()
	h.ServeError(response, customMessage)

	// Check status code
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusInternalServerError from ErrorHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the default message
	bodyString = response.Body.String()

	if bodyString != "Error: " + defaultMessage {
		t.Errorf("Expected \"Error: " + defaultMessage + 
			"\" from ErrorHandler. Got: %s", bodyString)
	}

	// Test AlwaysServeError with display errors off
	response = httptest.NewRecorder()
	h.AlwaysServeError(response, customMessage)

	// Check status code
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusInternalServerError from ErrorHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the custom message
	bodyString = response.Body.String()

	if bodyString != "Error: " + customMessage {
		t.Errorf("Expected \"Error: " + customMessage + 
			"\" from ErrorHandler. Got: %s", bodyString)
	}
}

// Test NotFoundHandler functions and methods
func TestNotFoundHandler(t *testing.T) {

	var(
		h *NotFoundHandler
		templatePath string
		bodyString string
		response *httptest.ResponseRecorder
		request *http.Request
	)

	// Set the paths
	templatePath = filepath.FromSlash("templates/notfound.html")

	// Get a NotFoundHandler with the not found template
	h = LoadNotFoundHandler(templatePath)

	// Test ServeHTTP with an arbitrary path
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/path", nil)
	h.ServeHTTP(response, request)

	// Check status code
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound from NotFoundHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the path
	bodyString = response.Body.String()

	if bodyString != "Not Found: /path" {
		t.Errorf("Expected \"Not Found: /path" + 
			"\" from NotFoundHandler. Got: %s", bodyString)
	}
}

// Test FileHandler functions and methods
func TestFileHandler(t *testing.T) {

	const(
		testFileBody string = "Test"
		sub1FileBody string = "Sub1"
		sub2FileBody string = "Sub2"
	)

	var(
		h *FileHandler
		nfh *NotFoundHandler
		templatePath string
		bodyString string
		location string
		response *httptest.ResponseRecorder
		request *http.Request
	)

	// Get a NotFoundHandler with the not found template
	templatePath = filepath.FromSlash("templates/notfound.html")
	nfh = LoadNotFoundHandler(templatePath)

	// Get a FileHandler on the test directory for the path "/test/"
	h = NewFileHandler("/test/", "./test", nfh)

	// Test ServeHTTP on "/test" without the trailing slash
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test", nil)
	h.ServeHTTP(response, request)

	// Check status code for found
	if response.Code != http.StatusFound {
		t.Errorf("Expected StatusFound from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response header contains the redirect to "/test/"
	location = response.HeaderMap["Location"][0]

	if location != "/test/" {
		t.Errorf("Expected a redirect to \"/test/\" from FileHandler. Got: %s", 
			response.HeaderMap)
	}

	// Test ServeHTTP on "/test/"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/", nil)
	h.ServeHTTP(response, request)

	// Check status code for ok
	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the contents of /test/index.html
	bodyString = response.Body.String()

	if bodyString != testFileBody {
		t.Errorf("Expected \"" + testFileBody + 
			"\" from FileHandler. Got: %s", bodyString)
	}

	// Test ServeHTTP on "/test/index.html"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/index.html", nil)
	h.ServeHTTP(response, request)

	// Check status code for moved permanently
	if response.Code != http.StatusMovedPermanently {
		t.Errorf("Expected StatusMovedPermanently from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response header contains the redirect to the test directory
	location = response.HeaderMap["Location"][0]

	if location != "./" {
		t.Errorf("Expected a redirect to \"./\" from FileHandler. Got: %s", 
			response.HeaderMap)
	}

	// Test ServeHTTP on "/test/sub1" without the trailing slash
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub1", nil)
	h.ServeHTTP(response, request)

	// Check status code for found
	if response.Code != http.StatusFound {
		t.Errorf("Expected StatusFound from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response header contains the redirect to "/test/sub1/"
	location = response.HeaderMap["Location"][0]

	if location != "/test/sub1/" {
		t.Errorf("Expected a redirect to \"/test/sub1/\" from FileHandler. Got: %s", 
			response.HeaderMap)
	}

	// Test ServeHTTP on "/test/sub1/"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub1/", nil)
	h.ServeHTTP(response, request)

	// Check status code for ok
	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the contents of /test/sub1/index.html
	bodyString = response.Body.String()

	if bodyString != sub1FileBody {
		t.Errorf("Expected \"" + sub1FileBody + 
			"\" from FileHandler. Got: %s", bodyString)
	}

	// Test ServeHTTP on "/test/sub1/index.html"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub1/index.html", nil)
	h.ServeHTTP(response, request)

	// Check status code for moved permamnently
	if response.Code != http.StatusMovedPermanently {
		t.Errorf("Expected StatusMovedPermanently from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response header contains the redirect to the sub1 directory
	location = response.HeaderMap["Location"][0]

	if location != "./" {
		t.Errorf("Expected a redirect to \"./\" from FileHandler. Got: %s", 
			response.HeaderMap)
	}	

	// Test ServeHTTP on "/test/sub2" without the trailing slash
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub2", nil)
	h.ServeHTTP(response, request)

	// Check status code for found
	if response.Code != http.StatusFound {
		t.Errorf("Expected StatusFound from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response header contains the redirect to "/test/sub1/"
	location = response.HeaderMap["Location"][0]

	if location != "/test/sub2/" {
		t.Errorf("Expected a redirect to \"/test/sub2/\" from FileHandler. Got: %s", 
			response.HeaderMap)
	}

	// Test ServeHTTP on "/test/sub2/"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub2/", nil)
	h.ServeHTTP(response, request)

	// Check status code for not found
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the path
	bodyString = response.Body.String()

	if bodyString != "Not Found: /test/sub2/" {
		t.Errorf("Expected \"Not Found: /test/sub2/" +
			"\" from FileHandler. Got: %s", bodyString)
	}	

	// Test ServeHTTP on "/test/sub2/not-index.html"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/sub2/not-index.html", nil)
	h.ServeHTTP(response, request)

	// Check status code for ok
	if response.Code != http.StatusOK {
		t.Errorf("Expected StatusOK from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the contents of /test/sub/not-index.html
	bodyString = response.Body.String()

	if bodyString != sub2FileBody {
		t.Errorf("Expected \"" + sub2FileBody + 
			"\" from FileHandler. Got: %s", bodyString)
	}

	// Test ServeHTTP on an arbitrary non existent file under "/test/"
	response = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/test/nofile", nil)
	h.ServeHTTP(response, request)

	// Check status code for not found
	if response.Code != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound from FileHandler. Got: %s", 
			response.Code)
	}

	// Check the response body contains the path
	bodyString = response.Body.String()

	if bodyString != "Not Found: /test/nofile" {
		t.Errorf("Expected \"Not Found: /test/nofile" +
			"\" from FileHandler. Got: %s", bodyString)
	}	
}