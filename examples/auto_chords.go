package examples

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"

	"github.com/rwelin/aujo"
)

type intervals []float64

func (i intervals) Equal(j intervals) bool {
	for k := range i {
		if i[k] != j[k] {
			return false
		}
	}
	return true
}

func (i intervals) Add(g float64) intervals {
	var ret intervals
	for _, f := range i {
		ret = append(ret, f+g)
	}
	return ret
}

func (i intervals) Expand() intervals {
	var exp []float64
	for _, f := range i {
		exp = append(exp, f, f-12, f+12, f-24)
	}
	sort.Float64s(exp)
	return intervals(exp)
}

var (
	s_Major = intervals{0, 2, 4, 5, 7, 9, 11}
	s_Minor = intervals{0, 2, 3, 5, 7, 8, 10, 11}
)

var (
	c_I          = intervals{0, 4, 7}
	c_iii7       = intervals{2, 4, 7, 11}
	c_vi7        = intervals{-3, 0, 4, 7}
	c_I6         = intervals{0, 4, 7, 9}
	c_ii7        = intervals{0, 2, 5, 9}
	c_IV7        = intervals{0, 4, 5, 9}
	c_V7         = intervals{2, 5, 7, 11}
	c_vii_dim_7  = intervals{-1, 2, 5, 9}
	c_ii7b5      = intervals{0, 2, 5, 8}
	c_bIII_aug_7 = intervals{2, 3, 7, 11}
	c_V7sus4     = intervals{2, 5, 7, 12}

	c_i        = intervals{0, 3, 7}
	c_ii_dim_7 = intervals{0, 2, 5, 8}
	c_III_7    = intervals{2, 3, 7, 10}
	c_iv7      = intervals{0, 3, 5, 8}
	c_VI7      = intervals{0, 3, 7, 8}
	c_vii_7    = intervals{2, 5, 8, 10}
)

var tonicChords = []intervals{
	c_I,
}

var subDominantChords = []intervals{
	c_ii7,
	c_IV7,
}

var dominantChords = []intervals{
	c_V7,
	c_vii_dim_7,
}

var otherChords = []intervals{
	c_iii7,
	c_vi7,
}

var minorTonicChords = []intervals{
	c_i,
}

var minorSubDominantChords = []intervals{
	c_ii_dim_7,
	c_iv7,
}

var minorDominantChords = []intervals{
	c_V7,
	c_vii_dim_7,
}

var minorOtherChords = []intervals{
	c_III_7,
	c_VI7,
	c_vii_7,
}

const (
	funcTonic = iota
	funcSubDom
	funcDom
	funcOther

	funcMax
)

var functionMap = map[int][]intervals{
	funcTonic:  tonicChords,
	funcSubDom: subDominantChords,
	funcDom:    dominantChords,
	funcOther:  otherChords,
}

var minorFunctionMap = map[int][]intervals{
	funcTonic:  minorTonicChords,
	funcSubDom: minorSubDominantChords,
	funcDom:    minorDominantChords,
	funcOther:  minorOtherChords,
}

func substituteChords(minor bool, function int) []intervals {
	if minor {
		return nil
	} else {
		switch function {
		case funcTonic:
			return []intervals{
				c_vi7,
				c_I6,
			}
		case funcDom:
			return []intervals{
				c_ii7b5,
				c_bIII_aug_7,
				c_V7sus4,
			}

		}
	}
	return nil
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

type nextChordResult struct {
	Function            int
	MinDistChord        intervals
	CanModulateUp       bool
	CanModulateDown     bool
	CanModulateRelative bool
}

func (a *autoChord) nextChord(prevFunction int, prevChord intervals, exclude []intervals) nextChordResult {

	scale := s_Major.Add(a.TonicPitch).Expand()
	if a.Minor {
		scale = s_Minor.Add(a.TonicPitch).Expand()
	}
	var chromaticNote float64
	var requiredNotes []float64
	for i := range prevChord {
		found := false
		for j := range scale {
			if prevChord[i] == scale[j] {
				found = true
				break
			} else if prevChord[i] < scale[j] {
				break
			}
		}
		if !found {
			chromaticNote = prevChord[i]
			break
		}
	}

	if chromaticNote != 0 {
		r := float64(0)
		for i := len(scale) - 1; i >= 0; i-- {
			if r == 0 ||
				math.Abs(scale[i]-chromaticNote) < math.Abs(r-chromaticNote) {
				r = scale[i]
			}
		}
		requiredNotes = append(requiredNotes, r)
	}

	for attempts := 0; ; attempts++ {
		function := nextFunction(prevFunction)
		chords := functionMap[function]
		if a.Minor {
			chords = minorFunctionMap[function]
		}
		if a.Modulate != modulationChance {
			chords = append(chords, substituteChords(a.Minor, function)...)
		}
		c := chords[rand.Intn(len(chords))]
		minDistChord := findMinDistChord(prevChord, c.Add(a.TonicPitch), requiredNotes)

		inExclude := false
		for _, e := range exclude {
			eq := true
			for i := range minDistChord {
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

		foundRequiredNote := len(requiredNotes)
		if foundRequiredNote > 0 {
			for _, r := range requiredNotes {
				for _, n := range minDistChord {
					if n == r {
						foundRequiredNote--
						break
					}
				}
			}
		}

		if (foundRequiredNote == 0 && !inExclude) ||
			(attempts > 50 && (foundRequiredNote == 0 || !inExclude)) ||
			attempts > 100 {
			return nextChordResult{
				Function:     function,
				MinDistChord: minDistChord,
				//CanModulateUp:   c.Equal(c_vi7),
				//CanModulateDown: c.Equal(c_ii7),
				//CanModulateUp: c.Equal(c_vi7) || c.Equal(c_I7) || c.Equal(c_iii7),
				CanModulateDown:     c.Equal(c_ii7) || c.Equal(c_IV7) || c.Equal(c_vi7),
				CanModulateRelative: a.Minor && c.Equal(c_vii_7) || !a.Minor && c.Equal(c_V7),
			}
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

		events = append(events, e)
	}
	return events
}

func findMinDistChord(prevChord intervals, nextChord intervals, requiredNotes []float64) intervals {
	exp := nextChord.Expand()

	const numChordNotes = 4

	minDist := float64(0)
	var minDistChord []float64
	for i := 0; i <= len(exp)-numChordNotes; i++ {
		c1 := exp[i : i+numChordNotes]
		var dist float64
		for j := 0; j < len(c1); j++ {
			d := 10 + math.Pow(prevChord[j]-c1[j], 2) + math.Pow((c1[j]-57)/5, 2)
			dist += math.Sqrt(d)
		}
		for j := 0; j < len(requiredNotes); j++ {
			found := false
			for k := 0; k < len(c1); k++ {
				if c1[k] == requiredNotes[j] {
					found = true
				}
			}
			if !found {
				dist += 100
			}
		}
		if minDistChord == nil || dist < minDist {
			minDist = dist
			minDistChord = c1
		}
	}

	return minDistChord
}

const chordDuration = 96000

const microOffset = 0.5

func (a *autoChord) progression(reps int, length int, prevBass float64, prevFunction int, prevChord intervals) (es []aujo.Event, lastBass float64, lastFunction int, lastChord intervals) {

	var cc []nextChordResult

	for i := 0; i < length; i++ {
		excludeChords := a.LastProgression
		for _, c := range cc {
			excludeChords = append(excludeChords, c.MinDistChord)
		}

		res := a.nextChord(prevFunction, prevChord, excludeChords)

		if i%2 == 1 {
			res.MinDistChord = res.MinDistChord.Add(microOffset)
		}

		cc = append(cc, res)
		prevChord = res.MinDistChord
		prevFunction = res.Function
	}

	a.LastProgression = a.LastProgression[:0]
	for _, c := range cc {
		a.LastProgression = append(a.LastProgression, c.MinDistChord)
	}

	var prevPrevBass float64

	fmt.Fprintln(os.Stderr, "TONIC", a.TonicPitch)
	for i := 0; i < reps; i++ {
		for j := 0; j < len(cc); j++ {
			c := cc[j].MinDistChord
			f := cc[j].Function
			if i == reps-1 && j == len(cc)-1 {
				penultimateChord := cc[j-1]
				if (penultimateChord.CanModulateUp || penultimateChord.CanModulateDown) &&
					rand.Intn(a.Modulate) == 0 {
					if penultimateChord.CanModulateUp {
						fmt.Fprintln(os.Stderr, "MODULATE UP")
						a.TonicPitch += 7
						if a.TonicPitch > 69 {
							a.TonicPitch -= 12
						}
					} else if penultimateChord.CanModulateDown {
						fmt.Fprintln(os.Stderr, "MODULATE DOWN")
						a.TonicPitch -= 7
						if a.TonicPitch < 69 {
							a.TonicPitch += 12
						}
					}
					c = findMinDistChord(penultimateChord.MinDistChord, c_V7.Add(a.TonicPitch), nil)
					f = funcDom
					prevChord = c
					prevFunction = funcDom
					a.Modulate = modulationChance
				} else if cc[j].CanModulateRelative && rand.Intn(a.Modulate) == 0 {
					if a.Minor {
						fmt.Fprintln(os.Stderr, "MODULATE RELATIVE MAJOR")
						a.Minor = false
						a.TonicPitch += 3
						if a.TonicPitch > 69 {
							a.TonicPitch -= 12
						}
					} else {
						fmt.Fprintln(os.Stderr, "MODULATE RELATIVE MINOR")

						a.Minor = true
						a.TonicPitch -= 3
						if a.TonicPitch < 66 {
							a.TonicPitch += 12
						}
					}
					prevFunction = funcDom
					a.Modulate = modulationChance
				}
			}

			var bass, bass1 float64
			for {
				getBassNote := func(prev float64, prevPrev float64) float64 {
					for {
						n := c[rand.Intn(len(c))]
						for n > 45 {
							n -= 12
						}
						for n < 29 {
							n += 12
						}
						if n != prev && n != prevPrev {
							return n
						}
					}
				}
				bass = getBassNote(prevBass, prevPrevBass)

				if bass-prevBass <= 3 && bass-prevBass > 0 {
					bass1 = bass - 1
				} else if prevBass-bass < 3 && prevBass-bass > 0 {
					bass1 = bass + 1
				} else {
					bass1 = getBassNote(bass, prevBass)
				}
				prevBass, prevPrevBass = bass, prevBass
				break
			}

			chordTime := int64(i*len(cc)+j) * chordDuration

			if bass1 != 0 {
				if i%2 == 0 {
					bass1 -= microOffset
				}

				walkingBassTime := chordTime - chordDuration/2
				es = append(es, events([]float64{bass1}, walkingBassTime)...)
			}

			fmt.Fprintln(os.Stderr, bass1, bass, c, f)
			es = append(es, events(append([]float64{bass}, c...), chordTime)...)
		}
	}
	fmt.Fprintln(os.Stderr)

	if a.Modulate > 1 {
		a.Modulate--
	}

	return es, prevBass, prevFunction, prevChord
}

func (a *autoChord) seq(prevBass float64, prevFunction int, prevChord intervals) *aujo.Sequence {

	e, prevBass, prevFunction, prevChord := a.progression(2, 4, prevBass, prevFunction, prevChord)

	return &aujo.Sequence{
		Events: append(e, aujo.Event{
			Time: e[len(e)-1].Time + chordDuration,
			Func: func(m *aujo.Mix) {
				m.SetNextSequence(a.seq(prevBass, prevFunction, prevChord))
			},
		}),
	}
}

type autoChord struct {
	TonicPitch      float64
	LastProgression []intervals
	Modulate        int
	Minor           bool
}

const modulationChance = 2

func AutoChords() *aujo.Sequence {
	a := &autoChord{
		TonicPitch: 69,
		Modulate:   modulationChance,
		Minor:      false,
	}
	return a.seq(0, funcDom, c_V7.Add(a.TonicPitch-12))
}
