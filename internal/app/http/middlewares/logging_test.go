package middlewares

import (
	"bytes"
	"encoding/json" // Ensure this is present
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	// "github.com/gorilla/mux" // Removed as it's unused in the test file
)

// CustomResponseWriter to capture status code if needed, though not strictly necessary for this middleware test
type CustomResponseWriter struct {
	httptest.ResponseRecorder
	statusCode int
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
	crw.ResponseRecorder.WriteHeader(statusCode)
}

func TestLoggingMiddleware(t *testing.T) {
	// 1. Setup Logger with captured output
	var logOutput bytes.Buffer
	// Use a JSON handler for easy parsing of structured logs, even though the app might use tint in dev.
	// The middleware itself is agnostic to the specific handler.
	testLogger := slog.New(slog.NewJSONHandler(&logOutput, &slog.HandlerOptions{
		// AddSource: true, // Can be useful for debugging tests
	}))

	// 2. Create the logging middleware instance
	loggingMiddleware := NewLogging(testLogger)

	// 3. Create a mock next handler
	mockNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Example action: set status code
		io.WriteString(w, "Hello from mock handler")
	})

	// 4. Create a dummy request
	req, err := http.NewRequest("GET", "/testpath", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.RemoteAddr = "1.2.3.4:12345" // Set a dummy remote address

	// 5. Create a ResponseRecorder (or our custom one if we need to check status written by next handler)
	rr := httptest.NewRecorder()

	// 6. Create a router and apply the middleware
	// The middleware is a mux.MiddlewareFunc, so we need a router to use it in a way it's intended.
	// However, we can also call the middleware directly for a simpler unit test.
	// The middleware function returns an http.Handler.
	// func(next http.Handler) http.Handler
	handlerUnderTest := loggingMiddleware(mockNextHandler)

	// 7. Serve the request
	start := time.Now() // To roughly estimate duration if needed, though the middleware calculates its own.
	handlerUnderTest.ServeHTTP(rr, req)
	end := time.Now()
	elapsed := end.Sub(start)

	// 8. Assertions
	// Check if the mockNextHandler was called (e.g., by checking response body or status)
	if status := rr.Code; status != http.StatusOK {
		// This check is more about the mock handler working than the middleware itself,
		// but good for sanity.
		t.Errorf("mockNextHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expectedBody := "Hello from mock handler"
	if rr.Body.String() != expectedBody {
		t.Errorf("mockNextHandler returned unexpected body: got %s want %s", rr.Body.String(), expectedBody)
	}

	// Check the log output
	logString := logOutput.String()
	t.Logf("Captured log: %s", logString) // Print log for debugging

	// Verify essential log fields
	if !strings.Contains(logString, `"msg":"request"`) && !strings.Contains(logString, `"message":"request"`) { // slog key is "msg" by default
		t.Errorf("log output does not contain 'request' message: %s", logString)
	}
	if !strings.Contains(logString, `"method":"GET"`) {
		t.Errorf("log output does not contain correct method: %s", logString)
	}
	if !strings.Contains(logString, `"path":"/testpath"`) {
		t.Errorf("log output does not contain correct path: %s", logString)
	}
	if !strings.Contains(logString, `"remote_addr":"1.2.3.4:12345"`) {
		t.Errorf("log output does not contain correct remote_addr: %s", logString)
	}
	if !strings.Contains(logString, `"duration":`) {
		t.Errorf("log output does not contain duration: %s", logString)
	}

	// Optional: More precise check on duration if necessary, e.g., parsing JSON
	// and checking if duration is positive and reasonable.
	// For example, check if it's less than the test's own elapsed time.
	// This requires parsing the JSON log. Example (simplified):
	// To ensure "encoding/json" is imported:
	var logEntry map[string]interface{}
	_ = json.Unmarshal([]byte{}, &logEntry) // Ensure import if needed by other logic

	if err := json.Unmarshal([]byte(logString), &logEntry); err == nil {
		if durationVal, ok := logEntry["duration"].(float64); ok { // Expect float64 for JSON numbers
			loggedDuration := time.Duration(int64(durationVal)) // Convert to time.Duration
			if loggedDuration < 0 || loggedDuration > elapsed+(50*time.Millisecond) { // Allow some buffer
				t.Errorf("Logged duration '%s' seems unreasonable compared to test elapsed time '%s'", loggedDuration, elapsed)
			}
		} else {
			t.Errorf("Logged duration is not a number (float64) in JSON: value is %T %v", logEntry["duration"], logEntry["duration"])
		}
	} else {
		t.Logf("Could not parse log output as JSON for duration check: %v. Log: %s", err, logString)
	}
}
