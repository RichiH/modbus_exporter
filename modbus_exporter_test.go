package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lupoDharkael/modbus_exporter/config"
)

func TestScrapeHandler(t *testing.T) {
	tests := []struct {
		name   string
		code   int
		config func() config.Config
		params map[string]string
	}{
		{
			name: "no module",
			code: http.StatusBadRequest,
			config: func() config.Config {
				return config.Config{}
			},
			params: map[string]string{},
		},
		{
			name: "no target",
			code: http.StatusBadRequest,
			config: func() config.Config {
				c := config.Config{}
				c.Modules = []config.Module{
					{
						Name: "my_module",
					},
				}

				return c
			},
			params: map[string]string{"module": "my_module"},
		},
		{
			name: "no sub_target",
			code: http.StatusBadRequest,
			config: func() config.Config {
				c := config.Config{}
				c.Modules = []config.Module{
					{
						Name: "my_module",
					},
				}

				return c
			},
			params: map[string]string{"module": "my_module", "target": "10.0.0.10"},
		},
		{
			name: "module and target",
			// The exporter won't be able to access the target,
			// thus, validation should pass (no 400) but scrape should
			// fail (500). One could stub the exporter itself.
			code: http.StatusInternalServerError,
			config: func() config.Config {
				c := config.Config{}
				c.Modules = []config.Module{
					{
						Name: "my_module",
					},
				}

				return c
			},
			params: map[string]string{"module": "my_module", "target": "10.0.0.10", "sub_target": "10"},
		},
	}

	for _, loopTest := range tests {
		test := loopTest

		t.Run(test.name, func(t *testing.T) {
			config := test.config()
			req, err := http.NewRequest("GET", "/metrics", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			for k, v := range test.params {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()

			scrapeHandler(config, rr, req)

			if status := rr.Code; status != test.code {
				t.Errorf(
					"handler returned wrong status code: got %v want %v, body: '%v'",
					status, test.code, rr.Body.String(),
				)
			}
		})
	}
}
