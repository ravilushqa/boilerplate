package httptest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type APIScenario struct {
	Name           string
	Method         string
	URL            string
	Body           io.Reader
	RequestHeaders map[string]string

	// Delay adds a delay before checking the expectations usually
	// to ensure that all fired non-awaited go routines have finished
	Delay time.Duration

	// expectations
	// ---
	ExpectedStatus     int
	ExpectedContent    []string
	NotExpectedContent []string

	// test hooks
	// ---
	Handler http.Handler
}

// Test executes the test case/scenario.
func (scenario *APIScenario) Test(t *testing.T) {
	if scenario.Handler == nil {
		t.Fatal("no handler provided")
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(scenario.Method, scenario.URL, scenario.Body)

	// set default header
	req.Header.Set("Content-Type", "application/json")

	// set scenario headers
	for k, v := range scenario.RequestHeaders {
		req.Header.Set(k, v)
	}

	scenario.Handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	var prefix = scenario.Name
	if prefix == "" {
		prefix = fmt.Sprintf("%s:%s", scenario.Method, scenario.URL)
	}

	if res.StatusCode != scenario.ExpectedStatus {
		t.Errorf("[%s] Expected status code %d, got %d", prefix, scenario.ExpectedStatus, res.StatusCode)
	}

	if scenario.Delay > 0 {
		time.Sleep(scenario.Delay)
	}

	if len(scenario.ExpectedContent) == 0 && len(scenario.NotExpectedContent) == 0 {
		if len(recorder.Body.Bytes()) != 0 {
			t.Errorf("[%s] Expected empty body, got \n%v", prefix, recorder.Body.String())
		}
	} else {
		// normalize json response format
		buffer := new(bytes.Buffer)
		err := json.Compact(buffer, recorder.Body.Bytes())
		var normalizedBody string
		if err != nil {
			// not a json...
			normalizedBody = recorder.Body.String()
		} else {
			normalizedBody = buffer.String()
		}

		for _, item := range scenario.ExpectedContent {
			if !strings.Contains(normalizedBody, item) {
				t.Errorf("[%s] Cannot find %v in response body \n%v", prefix, item, normalizedBody)
				break
			}
		}

		for _, item := range scenario.NotExpectedContent {
			if strings.Contains(normalizedBody, item) {
				t.Errorf("[%s] Didn't expect %v in response body \n%v", prefix, item, normalizedBody)
				break
			}
		}
	}
}
