package aujo

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"os"
	"sync"
)

const SamplingInterval = 2.0 * math.Pi / 44100.0

type Envelope struct {
	Value float64
	Time  int64
}

type Instrument struct {
	Harmonics []float64
	Attack    Envelope
	Decay     Envelope
	Sustain   Envelope
	Release   Envelope
}

func interpolate(index int64, e1 Envelope, val float64) float64 {
	lev := (e1.Value-val)/float64(e1.Time)*float64(index) + val
	if lev < 1e-10 {
		lev = 0
	}
	return lev
}

func (inst *Instrument) Level(event EventType, index int64, level float64) (float64, bool) {
	if index < 0 {
		return 0, false
	}

	switch event {
	case EventOn:
		if index < inst.Attack.Time {
			return interpolate(index, inst.Attack, level), true
		}
		index -= inst.Attack.Time
		if index < inst.Decay.Time {
			return interpolate(index, inst.Decay, inst.Attack.Value), true
		}
		index -= inst.Decay.Time
		if index < inst.Sustain.Time {
			return interpolate(index, inst.Sustain, inst.Decay.Value), true
		}
		index -= inst.Sustain.Time
		if index < inst.Release.Time {
			return interpolate(index, inst.Release, inst.Sustain.Value), true
		}
	case EventOff:
		if index < inst.Release.Time {
			return interpolate(index, inst.Release, level), true
		}
	}

	return 0, false
}

func (inst *Instrument) Mix(step float64) float64 {
	var sum float64
	for i, v := range inst.Harmonics {
		sum += v * math.Sin(step*float64(i+1))
	}
	return sum
}

type Channel struct {
	Event      EventType
	EventTime  int64
	EventLevel float64
	Pitch      float64
	PrevLevel  float64
}

type Voice struct {
	Level       float64
	Instrument  int
	VibratoFreq float64
	VibratoAmp  float64

	channels []Channel
}

type Mix struct {
	mutex sync.Mutex

	index    int64 // index is the current time
	seqIndex int64 // index in the current sequence

	rem  []byte      // rem are the bytes that are ready to be read
	bufC chan []byte // bufC is used to send the bytes to be read

	seq     *Sequence // seq is the currently playing sequence
	event   int       // event is the index of the next event
	nextSeq *Sequence // nextSeq is played after the current sequence has finished

	Level       float64 // master audio level
	Instruments []Instrument
	Voices      []Voice
}

var wavHeader = []byte{
	'R', 'I', 'F', 'F', // ChunkID
	0, 0, 0, 0, // ChunkSize
	'W', 'A', 'V', 'E', // Format
	'f', 'm', 't', ' ', // Subchunk1ID
	0x10, 0x0, 0x0, 0x0, // Subchunk1Size PCM
	0x1, 0x0, // AudioFormat PCM
	0x1, 0x0, // NumChannels Mono
	0x44, 0xAC, 0x0, 0x0, // SampleRate 44100Hz
	0x88, 0x58, 0x1, 0x0, // ByteRate 44100 * 1 * 16/8
	0x2, 0x0, // BlockAlign
	0x10, 0x0, // BitsPerSample
	'd', 'a', 't', 'a', // Subchunk2ID
	0xFF, 0xFF, 0xFF, 0xFF, // Subchunk2Size
}

func NewMix() *Mix {
	return &Mix{
		bufC: make(chan []byte),
		rem:  wavHeader,
	}
}

func ReadMixConfig(filename string) *Mix {
	m := NewMix()

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(f).Decode(m); err != nil {
		panic(err)
	}

	return m
}

func (m *Mix) Lock() {
	m.mutex.Lock()
}

func (m *Mix) Unlock() {
	m.mutex.Unlock()
}

func (m *Mix) Play(out io.Writer) {
	go m.Mix()

	for {
		io.CopyN(out, m, 1024)
	}
}

func (m *Mix) Read(buf []byte) (int, error) {
	if len(m.rem) == 0 {
		m.rem = <-m.bufC
	}
	n := copy(buf, m.rem)
	m.rem = m.rem[n:]
	return n, nil
}

func pitchToFreq(pitch float64) float64 {
	return 440.0 * math.Exp2((pitch-69)/12)
}

func (m *Mix) fill(buf []byte) int {
	m.Lock()
	defer m.Unlock()

	if m.seqIndex == 0 {
		m.seq = m.nextSeq
		m.event = 0
	}

	i := 0
	for ; i < len(buf); i += 2 {
		for {
			if len(m.seq.Events) == 0 {
				break
			}
			e := m.seq.Events[m.event]
			if e.Time > m.seqIndex {
				break
			}

			switch e.Type {
			case EventOn:
				fallthrough
			case EventOff:
				pitch := e.Pitch
				if e.PitchFunc != nil {
					pitch = e.PitchFunc()
				}
				v := &m.Voices[e.Voice]
				if pitch != 0 {
					var channel *Channel
					for i := range v.channels {
						if math.Abs(v.channels[i].Pitch-pitch) < 1e-2 {
							channel = &v.channels[i]
							break
						}
					}
					eventTime := m.index
					if channel == nil {
						v.channels = append(v.channels, Channel{
							Pitch:     pitch,
							Event:     e.Type,
							EventTime: eventTime,
						})
					} else {
						channel.Event = e.Type
						channel.EventTime = eventTime
						channel.EventLevel = channel.PrevLevel
					}
				}
			}
			if e.Func != nil {
				e.Func(m)
			}
			m.event++
			if m.event >= len(m.seq.Events) {
				m.event = 0
				m.seqIndex = 0
			}
		}

		s := float64(m.index) * SamplingInterval
		var sum float64
		for i, v := range m.Voices {
			vib := v.VibratoAmp * math.Sin(v.VibratoFreq*s)
			cs := v.channels[:0]
			for j, c := range v.channels {
				f := pitchToFreq(c.Pitch)
				offset := m.index - c.EventTime
				level, ok := m.Instruments[v.Instrument].Level(c.Event, offset, c.EventLevel)
				if ok {
					cs = append(cs, c)
					v.channels[j].PrevLevel = level
					sum += level * v.Level * m.Instruments[v.Instrument].Mix((s+vib)*f)
				}
			}
			m.Voices[i].channels = cs
		}
		m.index++
		m.seqIndex++

		t := uint16(m.Level * sum)
		binary.LittleEndian.PutUint16(buf[i:i+2], t)
	}
	return i
}

func (m *Mix) Mix() {
	for {
		buf := make([]byte, 1024)
		n := m.fill(buf)
		m.bufC <- buf[:n]
	}
}

func (m *Mix) SetNextSequence(s *Sequence) {
	m.Lock()
	defer m.Unlock()
	m.nextSeq = s
}

type EventType int

const (
	EventNone EventType = iota
	EventOn
	EventOff
)

type Event struct {
	Time      int64
	Pitch     float64
	PitchFunc func() float64
	Type      EventType
	Voice     int
	Func      func(*Mix)
}

type Sequence struct {
	Events []Event
}
