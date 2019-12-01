package examples

import "github.com/rwelin/aujo"

func Chords(scale []float64) *aujo.Sequence {
	return &aujo.Sequence{
		Events: []aujo.Event{
			{
				Time:  0,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[0] - 24,
			},
			{
				Time:  2000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[2] - 12,
			},
			{
				Time:  4000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[4] - 12,
			},
			{
				Time:  6000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 12,
			},

			{
				Time:  96000 + 0,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 36,
			},
			{
				Time:  96000 + 2000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[2] - 12,
			},
			{
				Time:  96000 + 4000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[4] - 12,
			},
			{
				Time:  96000 + 6000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 12,
			},

			{
				Time:  2*96000 + 0,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 36,
			},
			{
				Time:  2*96000 + 2000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[1] - 12,
			},
			{
				Time:  2*96000 + 4000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[4] - 12,
			},
			{
				Time:  2*96000 + 6000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 12,
			},

			{
				Time:  3*96000 + 0,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 36,
			},
			{
				Time:  3*96000 + 2000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[1] - 12,
			},
			{
				Time:  3*96000 + 4000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[3] - 12,
			},
			{
				Time:  3*96000 + 6000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[6] - 12,
			},

			{
				Time:  4*96000 + 48000 + 24000,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: scale[2] - 12,
			},

			{
				Time: 480000,
			},
		},
	}
}
