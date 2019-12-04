package examples

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"

	"github.com/rwelin/aujo"
)

var tonicChords = [][]float64{
	{0, 4, 7, 11}, // I7
	{2, 4, 7, 11}, // iii7
	{-3, 0, 4, 7}, // vi7
	{0, 4, 7, 9},  // I6
}

var subDominantChords = [][]float64{
	{0, 2, 5, 9}, // ii7
	{0, 4, 5, 9}, // IV7
}

var dominantChords = [][]float64{
	{2, 5, 7, 11}, // V7
	{-1, 2, 5, 9}, // vii-7
	{0, 2, 5, 8},  // ii7b5
	{2, 3, 7, 11}, // bIII+7
	{2, 5, 7, 12}, // V7sus4
}

var otherChords = [][]float64{
	{2, 4, 7, 11}, // iii7
	{-3, 0, 4, 7}, // vi7
}

const (
	funcTonic = iota
	funcSubDom
	funcDom
	funcOther

	funcMax
)

var functionMap = map[int][][]float64{
	funcTonic:  tonicChords,
	funcSubDom: subDominantChords,
	funcDom:    dominantChords,
	funcOther:  otherChords,
}

func middleFunctions(prevFunction int, nextFunction int) []int {
	pot := make(map[int]struct{})
	fromPrev := followingFunctions(prevFunction)
	for _, f := range fromPrev {
		p := followingFunctions(f)
		for _, g := range p {
			if g == nextFunction {
				pot[f] = struct{}{}
				break
			}
		}
	}

	var funcs []int
	for f := range pot {
		funcs = append(funcs, f)
	}
	return funcs
}

func followingFunctions(currentFunction int) []int {
	var funcs []int
	switch currentFunction {
	case funcTonic:
		funcs = []int{funcSubDom, funcDom, funcOther}
	case funcSubDom:
		funcs = []int{funcTonic, funcDom, funcOther}
	case funcDom:
		funcs = []int{funcTonic}
	case funcOther:
		funcs = []int{funcTonic, funcSubDom, funcDom}
	default:
		panic("no")
	}
	return funcs
}

func nextFunction(currentFunction int) int {
	funcs := followingFunctions(currentFunction)
	return funcs[rand.Intn(len(funcs))]
}

func (a *autoChord) chord(prevFunction int, prevChord []float64, exclude [][]float64) (int, []float64) {

	for attempts := 0; ; attempts++ {
		function := nextFunction(prevFunction)
		chords := functionMap[function]
		c := chords[rand.Intn(len(chords))]
		minDistChord := findMinDistChord(prevChord, a.TonicPitch, c)

		inExclude := false
		for _, e := range exclude {
			eq := true
			for i := range c {
				if e[i] != minDistChord[i] {
					eq = false
					break
				}
			}
			if eq {
				inExclude = true
				break
			}
		}
		if !inExclude || attempts > 100 {
			return function, minDistChord
		}
	}
}

func events(chord []float64, offset int64) []aujo.Event {
	var events []aujo.Event
	for i, f := range chord {
		e := aujo.Event{
			Time:  offset + 2000*int64(i),
			Voice: 3,
			Type:  aujo.EventOn,
			Pitch: f,
		}
		if i == 0 {
			e.Pitch -= 12
		} else if i == 1 {
			e1 := e
			e1.Pitch += 12
			events = append(events, e1)
		}
		events = append(events, e)
	}
	return events
}

func findMinDistChord(prevChord []float64, tonicPitch float64, nextChord []float64) []float64 {
	var exp []float64
	for _, f := range nextChord {
		f += tonicPitch
		exp = append(exp, f, f-12, f+12)
	}
	sort.Float64s(exp)

	minDist := float64(0)
	var minDistChord []float64
	for i := 0; i <= len(exp)-len(nextChord); i++ {
		c1 := exp[i : i+len(nextChord)]
		var dist float64
		for j := 0; j < len(c1); j++ {
			d := math.Abs(prevChord[j] - c1[j])
			switch j {
			case 0:
				// Make the base more likely to move
				if d < 2 {
					d += 2
				}
			case 1:
				if d < 1 {
					d += 2
				}
			}
			dist += d
		}
		if minDistChord == nil || dist < minDist {
			minDist = dist
			minDistChord = c1
		}
	}

	return minDistChord
}

const chordDuration = 96000

func (a *autoChord) progression(reps int, length int, prevFunction int, prevChord []float64) (es []aujo.Event, lastFunction int, lastChord []float64) {

	var cc [][]float64

	for i := 0; i < length; i++ {

		function, minDistChord := a.chord(prevFunction, prevChord, cc)
		cc = append(cc, minDistChord)
		prevFunction = function
		prevChord = minDistChord
	}

	for i := 0; i < reps; i++ {
		for j := 0; j < len(cc); j++ {
			c := cc[j]
			fmt.Fprintln(os.Stderr, c)
			es = append(es, events(c, int64((i*len(cc)+j)*chordDuration))...)
		}
	}
	fmt.Fprintln(os.Stderr)

	return es, prevFunction, prevChord
}

func (a *autoChord) seq(prevFunction int, prevChord []float64) *aujo.Sequence {

	e, prevFunction, prevChord := a.progression(2, 4, prevFunction, prevChord)

	return &aujo.Sequence{
		Events: append(e, aujo.Event{
			Time: e[len(e)-1].Time + chordDuration,
			Func: func(m *aujo.Mix) {
				m.SetNextSequence(a.seq(prevFunction, prevChord))
			},
		}),
	}
}

type autoChord struct {
	TonicPitch float64
}

func AutoChords() *aujo.Sequence {
	a := &autoChord{
		TonicPitch: 57,
	}
	return a.seq(funcTonic, tonicChords[0])
}
