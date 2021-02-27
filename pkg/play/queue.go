package play

import (
	"github.com/paynejacob/speakerbob/pkg/sound"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"sync"
	"time"
)

const messageType websocket.MessageType = "play"

type messagePayload struct {
	Scheduled time.Time   `json:"scheduled"`
	Sound     sound.Sound `json:"sound"`
}

type queue struct {
	m sync.RWMutex

	playChannel chan bool

	sounds []sound.Sound
}

func newQueue() *queue {
	return &queue{
		sync.RWMutex{},
		make(chan bool, 0),
		make([]sound.Sound, 0),
	}
}

func (q *queue) EnqueueSound(s sound.Sound) {
	q.m.Lock()
	defer q.m.Unlock()

	q.sounds = append(q.sounds, s)

	q.playChannel <- true
}

func (q *queue) ConsumeQueue(ws *websocket.Service) {
	var timer *time.Timer
	var isEmpty bool
	var isPlaying bool
	var _sound sound.Sound

	timer = time.NewTimer(0)

	for {
		select {
		case <-q.playChannel:
			// if something is already playing we do nothing
			if isPlaying {
				continue
			}

			// get the next sound off the queue and play it
			_sound, isEmpty = q.pop()
			ws.BroadcastMessage(messageType, messagePayload{
				Scheduled: time.Now(),
				Sound:     _sound,
			})
			isPlaying = true
			timer.Reset(_sound.Duration) // set a timer for the duration of the sound
		case <-timer.C:
			// the sound finished playing
			isPlaying = false

			// play the next sound, if there is no next sound exit
			_sound, isEmpty = q.pop()
			if isEmpty {
				continue
			}

			ws.BroadcastMessage(messageType, messagePayload{
				Scheduled: time.Now(),
				Sound:     _sound,
			})
			isPlaying = true
			timer.Reset(_sound.Duration) // set a timer for the duration of the sound
		}
	}
}

func (q *queue) empty() bool {
	q.m.RLock()
	defer q.m.RUnlock()

	return len(q.sounds) == 0
}

func (q *queue) pop() (s sound.Sound, empty bool) {
	empty = q.empty()

	if empty {
		return
	}

	q.m.RLock()
	defer q.m.RUnlock()

	s = q.sounds[0]
	q.sounds = q.sounds[1:]

	return
}
