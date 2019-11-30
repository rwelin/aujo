package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"

	"github.com/rwelin/aujo"
	"github.com/rwelin/aujo/api"
)

const SamplingInterval = 2.0 * math.Pi / 44100.0

var major = []float64{69, 71, 73, 74, 76, 78, 80}
var minor = []float64{69, 71, 72, 74, 76, 77, 79}
var melMinor = []float64{69, 71, 72, 74, 76, 78, 80}
var harMinor = []float64{69, 71, 72, 74, 75, 77, 80}

func play(m *aujo.Mix) {
	out := os.Stdout

	w := out.Write
	w([]byte("RIFF"))                 // ChunkID
	w([]byte{0, 0, 0, 0})             // ChunkSize
	w([]byte("WAVE"))                 // Format
	w([]byte("fmt "))                 // Subchunk1ID
	w([]byte{0x10, 0, 0, 0})          // Subchunk1Size PCM
	w([]byte{0x1, 0})                 // AudioFormat PCM
	w([]byte{0x1, 0})                 // NumChannels Mono
	w([]byte{0x44, 0xAC, 0x0, 0x0})   // SampleRate 44100Hz
	w([]byte{0x88, 0x58, 0x1, 0x0})   // ByteRate 44100 * 1 * 16/8
	w([]byte{0x2, 0x0})               // BlockAlign
	w([]byte{0x10, 0x0})              // BitsPerSample
	w([]byte("data"))                 // Subchunk2ID
	w([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // Subchunk2Size

	minorScale := false

	i := 0
	p := 0
	buf := make([]byte, 2)
	for {
		if i%6000 == 0 {
			d := rand.Intn(6) - 2
			scale := major
			if minorScale {
				if d < 0 {
					scale = minor
				} else {
					scale = harMinor
				}
			}
			p = p + d
			if p < 0 {
				p += len(scale)
			} else if p >= len(scale) {
				p -= len(scale)
			}

			m.Lock()
			m.Voices[0].Pitch = scale[p]
			m.Voices[1].Pitch = scale[(p+2)%len(scale)]
			m.Unlock()
		}

		if i%441000 == 0 {
			minorScale = !minorScale
		}

		i += m.Mix(buf, i)

		out.Write(buf)
	}
}

const ConfigFilename = "config.json"

func ReadConfig() *aujo.Mix {
	var m aujo.Mix
	f, err := os.Open(ConfigFilename)
	if err != nil {
		defaultInstrument := [20]float64{1.0}
		return &aujo.Mix{
			Level:            10000,
			SamplingInterval: SamplingInterval,
			Instruments: []aujo.Instrument{
				{
					Harmonics: defaultInstrument[:],
				},
			},
			Voices: []aujo.Voice{
				{
					Level:      0.4,
					Pitch:      69,
					Instrument: 0,
				},
				{
					Level:      0.1,
					Pitch:      69,
					Instrument: 0,
				},
			},
		}
	}

	if err := json.NewDecoder(f).Decode(&m); err != nil {
		panic(err)
	}

	return &m
}

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
	m := ReadConfig()

	go play(m)

	handler := api.NewHandler(&apiCallbacks{
		m: m,
	})

	panic(http.ListenAndServe(":7999", handler))
}
