package main

import (
	"encoding/json"
	"fmt"
	"io"
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

	for {
		io.CopyN(out, m, 1024)
	}
}

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
				Time: 0,
				Func: func(m *aujo.Mix) {
					scale := major
					p += rand.Intn(6) - 2
					if p < 0 {
						p += len(scale)
					} else if p >= len(scale) {
						p -= len(scale)
					}

					m.Voices[0].Pitch = scale[p]
				},
			},
			{
				Time: 12000,
				Func: func(m *aujo.Mix) {
					m.Voices[1].Pitch = scale[(p+2)%len(scale)]
				},
			},
			{
				Time: 24000,
			},
		},
	})

	go m.Mix()
	go play(m)

	handler := api.NewHandler(&apiCallbacks{
		m: m,
	})

	panic(http.ListenAndServe(":7999", handler))
}
