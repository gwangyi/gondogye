package main

import "flag"
import "log"
import "github.com/gwangyi/gondogye/dht"
import "github.com/gwangyi/gondogye/server"

var host = flag.String("host", "", "Listening host")
var port = flag.Int("port", 10000, "Listening port")
var pin = flag.Int("pin", -1, "GPIO pin number in BCM2835 way")
var verbose = flag.Int("v", 0, "Set log verbosity level. Log will be written to stderr")

var sensor dht.DHT

func main() {
	flag.Parse()
	if *pin < 0 {
		log.Fatalf("Pin number must be specified")
	}
	sensor = dht.NewDHT11(*pin)
	sensor.Verbose(*verbose)
	(&server.Server{Sensor: sensor}).Start(*host, *port)
}
