package dht

import "fmt"

// #cgo CFLAGS: -I.
// #include <sys/types.h>
// #include <bcm2835.h>
// #include <dht.h>
import "C"

func init() {
	if C.bcm2835_init() == 0 {
		panic("bcm2835 init failed")
	}
}

// Result represents the result of DHT sensor.
type Result struct {
	Humidity    float32
	Temperature float32
}

// DHT refers DHT sensor.
type DHT interface {
	Verbose(level int)
	Read() (Result, error)
}

// cDHT refers thin C wrapper of DHT sensor.
type cDHT interface {
	Verbose(level int)
	Read() (uint32, error)
}

func newDHT(pin int) *C.struct_dht_sensor {
	dht := C.struct_dht_sensor{}
	C.dht_init(&dht, C.int(pin))
	return &dht
}

// Verbose enables/disables DHT logger that writes everything to stderr.
func (dht *C.struct_dht_sensor) Verbose(level int) {
	C.dht_set_loglevel(dht, C.int(level))
}

// Read reads current value from DHT sensor.
func (dht *C.struct_dht_sensor) Read() (uint32, error) {
	output := C.uint32_t(0)
	ret := C.dht_read(dht, &output)
	if ret != 0 {
		return uint32(output), nil
	}
	return 0, fmt.Errorf("Unexpected behavior on the DHT sensor at %v", dht.pin)
}

type dhtImpl struct {
	cdht cDHT
}

type dht11 struct {
	dhtImpl
}

type dht22 struct {
	dhtImpl
}

func (dht *dhtImpl) Verbose(level int) {
	dht.cdht.Verbose(level)
}

// NewDHT11 creates new DHT sensor object that talks to DHT11 sensor.
func NewDHT11(pin int) DHT {
	return &dht11{dhtImpl: dhtImpl{cdht: newDHT(pin)}}
}

func (dht *dht11) Read() (Result, error) {
	output, err := dht.cdht.Read()
	if err != nil {
		return Result{}, err
	}
	return Result{
		Humidity:    float32((output >> 24) & 0xff),
		Temperature: float32((output >> 8) & 0xff),
	}, nil
}

// NewDHT22 creates new DHT sensor object that talks to DHT22 sensor.
func NewDHT22(pin int) DHT {
	return &dht22{dhtImpl: dhtImpl{cdht: newDHT(pin)}}
}

func (dht *dht22) Read() (Result, error) {
	output, err := dht.cdht.Read()
	if err != nil {
		return Result{}, err
	}
	hum := (output >> 16) & 0xffff
	temp := output & 0xffff
	return Result{
		Humidity:    float32(hum) / 10,
		Temperature: float32(temp) / 10,
	}, nil
}
