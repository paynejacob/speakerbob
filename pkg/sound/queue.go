package sound

import (
	"context"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"sync"
	"time"
)

type queue struct {
	m sync.RWMutex

	playChannel chan bool

	sounds []Sound
}

func (q *queue) EnqueueSounds(sounds ...Sound) {
	q.m.Lock()
	defer q.m.Unlock()

	for i := range sounds {
		q.sounds = append(q.sounds, sounds[i])
	}

	q.playChannel <- true
}

func (q *queue) ConsumeQueue(ctx context.Context, ws *websocket.Service) {
	var timer *time.Timer
	var isEmpty bool
	var isPlaying bool
	var _sound Sound

	timer = time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			break
		case <-q.playChannel:
			// if something is already playing we do nothing
			if isPlaying {
				continue
			}

			// get the next sound off the queue and play it
			_sound, _ = q.pop()
			ws.BroadcastMessage(PlayMessage{
				Type:      websocket.PlayMessageType,
				Sound:     _sound,
				Scheduled: time.Now(),
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

			ws.BroadcastMessage(PlayMessage{
				Type:      websocket.PlayMessageType,
				Sound:     _sound,
				Scheduled: time.Now(),
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

func (q *queue) pop() (s Sound, empty bool) {
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
