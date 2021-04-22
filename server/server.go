package server

import "encoding/json"
import "fmt"
import "log"
import "net/http"

import "github.com/gwangyi/gondogye/dht"

type Server struct {
	Sensor dht.DHT
}

func (s *Server) measure(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/json")
	result, err := s.Sensor.Read()
	if err != nil {
		fmt.Fprintf(w, "{\"err\": %q}", err.Error())
		log.Printf("Failed: %v", err)
		return
	}
	response, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(w, "{\"err\": %q}", err.Error())
		log.Printf("Failed: %v", err)
		return
	}
	w.Write(response)

	log.Printf("Measured: %#v", result)
}

func (s *Server) Start(host string, port int) {
	log.Printf("Listening from %v:%v", host, port)
	http.HandleFunc("/", s.measure)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%v", host, port), nil))
}
