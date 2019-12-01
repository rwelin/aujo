package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"

	"github.com/rwelin/aujo"
	"github.com/rwelin/aujo/api"
)

var major = []float64{69, 71, 73, 74, 76, 78, 80}
var minor = []float64{69, 71, 72, 74, 76, 77, 79}
var melMinor = []float64{69, 71, 72, 74, 76, 78, 80}
var harMinor = []float64{69, 71, 72, 74, 75, 77, 80}

const ConfigFilename = "config.json"

type apiCallbacks struct {
	m *aujo.Mix
}

func log(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func (cb *apiCallbacks) UpdateInstrumentHarmonics(inst int, harm []float64) error {
	config, err := ioutil.TempFile("", "aujoconfig")
	if err != nil {
		panic(err)
	}
	defer os.Remove(config.Name())
	defer config.Close()

	cb.m.Lock()
	defer cb.m.Unlock()

	if inst >= len(cb.m.Instruments) {
		return fmt.Errorf("no such instrument")
	}

	cb.m.Instruments[inst].Harmonics = harm

	if err := json.NewEncoder(config).Encode(cb.m); err != nil {
		panic(err)
	}
	if err := os.Rename(config.Name(), ConfigFilename); err != nil {
		panic(err)
	}

	return nil
}

func (cb *apiCallbacks) Mix() ([]byte, error) {
	cb.m.Lock()
	defer cb.m.Unlock()
	return json.Marshal(cb.m)
}

func main() {
	m := aujo.ReadMixConfig(ConfigFilename)

	p := 0
	scale := major

	m.SetNextSequence(&aujo.Sequence{
		Events: []aujo.Event{
			{
				Time:  0,
				Voice: 0,
				Type:  aujo.EventOn,
				PitchFunc: func() float64 {
					p += rand.Intn(6) - 2
					if p < 0 {
						p += len(scale)
					} else if p >= len(scale) {
						p -= len(scale)
					}
					return scale[p]
				},
			},
			{
				Time:  6000,
				Voice: 1,
				Type:  aujo.EventOn,
				PitchFunc: func() float64 {
					return scale[(p+2)%len(scale)]
				},
			},
			{
				Time:  10500,
				Voice: 0,
				Type:  aujo.EventOff,
				PitchFunc: func() float64 {
					return scale[p]
				},
			},
			{
				Time:  10000,
				Voice: 1,
				Type:  aujo.EventOff,
				PitchFunc: func() float64 {
					return scale[(p+2)%len(scale)]
				},
			},
			{
				Time: 12000,
			},
		},
	})

	go m.Play(os.Stdout)

	handler := api.NewHandler(&apiCallbacks{
		m: m,
	})

	panic(http.ListenAndServe(":7999", handler))
}
