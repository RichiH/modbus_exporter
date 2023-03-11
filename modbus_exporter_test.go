// Copyright 2019 Max Inden, Richard Hartmann, and The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RichiH/modbus_exporter/config"
	"github.com/RichiH/modbus_exporter/modbus"
	"github.com/go-kit/log"
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
			exporter := modbus.NewExporter(config)

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

			scrapeHandler(exporter, rr, req, log.NewNopLogger())

			if status := rr.Code; status != test.code {
				t.Errorf(
					"handler returned wrong status code: got %v want %v, body: '%v'",
					status, test.code, rr.Body.String(),
				)
			}
		})
	}
}
