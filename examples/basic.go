package examples

import (
	"math/rand"

	"github.com/rwelin/aujo"
)

func Basic(scale []float64) *aujo.Sequence {
	p := 0
	return &aujo.Sequence{
		Events: []aujo.Event{
			{
				Time:  0,
				Voice: 2,
				Type:  aujo.EventOn,
				Pitch: 35,
			},
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
				Time:  12000,
				Voice: 1,
				Type:  aujo.EventOn,
				PitchFunc: func() float64 {
					return scale[(p+2)%len(scale)]
				},
			},
			{
				Time:  20500,
				Voice: 0,
				Type:  aujo.EventOff,
				PitchFunc: func() float64 {
					return scale[p]
				},
			},
			{
				Time:  20000,
				Voice: 1,
				Type:  aujo.EventOff,
				PitchFunc: func() float64 {
					return scale[(p+2)%len(scale)]
				},
			},
			{
				Time: 24000,
			},
		},
	}
}
