package server

import "testing"
import "encoding/json"
import "net/http"
import "net/http/httptest"

import "github.com/gwangyi/gondogye/dht"
import "github.com/google/go-cmp/cmp"

type fakeDHT struct {
	err    error
	result *dht.Result
}

func (d *fakeDHT) Verbose(level int) {
}

func (d *fakeDHT) Read() (dht.Result, error) {
	if d.err != nil {
		return dht.Result{Humidity: 0, Temperature: 0}, d.err
	}
	return *d.result, nil
}

func TestMeasure(t *testing.T) {
	want := &dht.Result{Humidity: 40.0, Temperature: 20.0}

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	s := &Server{Sensor: &fakeDHT{result: want}}
	s.measure(res, req)
	if res.Result().StatusCode != 200 {
		t.Errorf("StatusCode %v != 200", res.Result().StatusCode)
	}

	contentTypes, ok := res.Result().Header[http.CanonicalHeaderKey("Content-Type")]
	if !ok || len(contentTypes) == 0 {
		t.Errorf("Content type header missing")
	} else if len(contentTypes) > 1 || contentTypes[0] != "text/json" {
		t.Errorf("Content-Type %#v != [\"text/json\"]", contentTypes)
	}

	got := &dht.Result{}
	decoder := json.NewDecoder(res.Result().Body)
	if err := decoder.Decode(got); err != nil {
		t.Error(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
