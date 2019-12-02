package examples

import "github.com/rwelin/aujo"

func AutoChords() *aujo.Sequence {
	s := &aujo.Sequence{
		Events: []aujo.Event{
			{
				Time:  0,
				Voice: 4,
				Type:  aujo.EventOn,
				Pitch: 57,
			},
			{
				Time:  0,
				Voice: 4,
				Type:  aujo.EventOn,
				Pitch: 61.0782,
			},
			{
				Time:  0,
				Voice: 3,
				Type:  aujo.EventOn,
				Pitch: 64.0196,
			},
			{
				Time: 96000,
			},
		},
	}

	return s
}
