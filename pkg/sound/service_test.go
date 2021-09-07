package sound

import (
	"bytes"
	"fmt"
	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/mux"
	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/paynejacob/speakerbob/pkg/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var soundProvider *SoundProvider
var groupProvider *GroupProvider
var websocketService websocket.Service
var maxDuration time.Duration
var playChannel chan bool
var svc *Service

func init() {
	websocketService = websocket.Service{}

	// because we don't run the consumer we need to unblock operations that block for the consumer
	playChannel = make(chan bool, 0)
	go func() {
		for {
			<- playChannel
		}
	}()

	maxDuration, _ = time.ParseDuration("10s")
}

func setup() {
	soundProvider = & SoundProvider{
		Store:       memory.New(),
	}
	_= soundProvider.Initialize()

	groupProvider = &GroupProvider{
		Store: memory.New(),
	}
	_= groupProvider.Initialize()
}

func newServer() *httptest.Server {
	svc = &Service{
		SoundProvider:    soundProvider,
		GroupProvider:    groupProvider,
		WebsocketService: &websocketService,
		MaxSoundDuration: maxDuration,
		playQueue:        playQueue{
			m:           sync.RWMutex{},
			playChannel: playChannel,
			sounds:      make([]Sound, 0),
		},
	}

	router := mux.NewRouter()
	svc.RegisterRoutes(router)

	return httptest.NewServer(router)
}

func TestListSound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	s3 := NewSound()
	s3.Name = "s3"
	s3.Hidden = false

	shidden := NewSound()

	// empty
	httpexpect.New(t, sut.URL).
		GET("/sound/sounds/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		Empty()

	// hidden
	_ = soundProvider.Save(&shidden)

	httpexpect.New(t, sut.URL).
		GET("/sound/sounds/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		Empty()

	// 1 sound
	_ = soundProvider.Save(&s1)

	httpexpect.New(t, sut.URL).
		GET("/sound/sounds/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		ContainsOnly(s1)

	// 3 sound
	_ = soundProvider.Save(&s2)
	_ = soundProvider.Save(&s3)

	httpexpect.New(t, sut.URL).
		GET("/sound/sounds/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		ContainsOnly(s1, s2, s3)
}

func TestCreateSound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	// invalid audio file
	httpexpect.New(t, sut.URL).
		POST("/sound/sounds/").
		WithMultipart().
		WithFileBytes("foo.mp3", "foo.mp3", make([]byte, 64)).
		Expect().
		Status(http.StatusNotAcceptable)

	// valid audio file
	body := []byte{82, 73, 70, 70, 36, 0, 0, 0, 87, 65, 86, 69, 102, 109, 116, 32, 16, 0, 0, 0, 1, 0, 1, 0, 68, 172, 0, 0, 136, 88, 1, 0, 2, 0, 16, 0, 100, 97, 116, 97, 0, 0, 0, 0}
	httpexpect.New(t, sut.URL).
		POST("/sound/sounds/").
		WithMultipart().
		WithFileBytes("foo.wav", "foo.wav", body).
		Expect().
 		Status(http.StatusCreated).
		JSON()
}

func TestUpdateSound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	sound := NewSound()
	sound.Name = "test"
	sound.Hidden = false

	_ = soundProvider.Save(&sound)

	body := sound


	// invalid sound id
	httpexpect.New(t, sut.URL).
		PATCH("/sound/sounds/foobar/").
		WithJSON(&body).
		Expect().
		Status(http.StatusNotFound)

	// invalid sound name
	body.Name = ""
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/sounds/%s/", sound.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusNotAcceptable)

	// valid
	body.Name = "test2"
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/sounds/%s/", sound.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusAccepted)
}

func TestDeleteSound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	s3 := NewSound()
	s3.Name = "s3"
	s3.Hidden = false

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id, s3.Id}

	g2 := NewGroup()
	g2.Name = "g2"
	g2.SoundIds = []string{s2.Id, s3.Id}

	// sound does not exist
	httpexpect.New(t, sut.URL).
		DELETE(fmt.Sprintf("/sound/sounds/%s/", s1.Id)).
		Expect().
		Status(http.StatusNoContent)

	// sound exists
	_ = soundProvider.Save(&s1)
	httpexpect.New(t, sut.URL).
		DELETE(fmt.Sprintf("/sound/sounds/%s/", s1.Id)).
		Expect().
		Status(http.StatusNoContent)

	assert.Len(t, soundProvider.List(), 0)

	// sound in group
	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)
	_ = soundProvider.Save(&s3)
	_ = groupProvider.Save(&g1)
	_ = groupProvider.Save(&g2)
	httpexpect.New(t, sut.URL).
		DELETE(fmt.Sprintf("/sound/sounds/%s/", s1.Id)).
		Expect().
		Status(http.StatusNoContent)

	assert.Len(t, soundProvider.List(), 2)
	assert.Len(t, groupProvider.List(), 1)
}

func TestPlaySound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	// play sound
	httpexpect.New(t, sut.URL).
		PUT(fmt.Sprintf("/sound/sounds/%s/play/", s1.Id)).
		Expect().
		Status(http.StatusAccepted)

	assert.Len(t, svc.playQueue.sounds, 1)

	// enqueue sound
	httpexpect.New(t, sut.URL).
		PUT(fmt.Sprintf("/sound/sounds/%s/play/", s2.Id)).
		Expect().
		Status(http.StatusAccepted)

	assert.Len(t, svc.playQueue.sounds, 2)
	assert.Equal(t, svc.playQueue.sounds[0].Id, s1.Id)
	assert.Equal(t, svc.playQueue.sounds[1].Id, s2.Id)

	// invalid sound id
	httpexpect.New(t, sut.URL).
		PUT(fmt.Sprintf("/sound/sounds/%s/play/", "foobar")).
		Expect().
		Status(http.StatusNotFound)
}

func TestDownloadSound(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	audioContent := []byte{1,2,3}

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.WriteAudio(&s1, bytes.NewReader(audioContent))

	// invalid sound
	httpexpect.New(t, sut.URL).
		GET(fmt.Sprintf("/sound/sounds/%s/download/", "foobar")).
		Expect().
		Status(http.StatusNotFound)

	// valid sound
	httpexpect.New(t, sut.URL).
		GET(fmt.Sprintf("/sound/sounds/%s/download/", s1.Id)).
		Expect().
		Status(http.StatusOK).
		ContentType("audio/mpeg").
		Body().
		Equal(string(audioContent))
}

func TestListGroup(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	s3 := NewSound()
	s3.Name = "s3"
	s3.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)
	_ = soundProvider.Save(&s3)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id,s2.Id,s3.Id}

	g2 := NewGroup()
	g2.Name = "g2"
	g2.SoundIds = []string{s3.Id,s2.Id,s1.Id}

	g3 := NewGroup()
	g3.Name = "g3"
	g3.SoundIds = []string{s3.Id,s2.Id,s3.Id}

	// empty
	httpexpect.New(t, sut.URL).
		GET("/sound/groups/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		Empty()

	// 1 group
	_ = groupProvider.Save(&g1)

	httpexpect.New(t, sut.URL).
		GET("/sound/groups/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		ContainsOnly(g1)

	// 3 sound
	_ = groupProvider.Save(&g2)
	_ = groupProvider.Save(&g3)

	httpexpect.New(t, sut.URL).
		GET("/sound/groups/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().
		ContainsOnly(g1, g2, g3)
}

func TestCreateGroup(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	s3 := NewSound()
	s3.Name = "s3"
	s3.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}

	// bad name
	g1.Name = ""
	httpexpect.New(t, sut.URL).
		POST("/sound/groups/").
		WithJSON(g1).
		Expect().
		Status(http.StatusNotAcceptable)

	// bad sound id
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s3.Id}
	httpexpect.New(t, sut.URL).
		POST("/sound/groups/").
		WithJSON(g1).
		Expect().
		Status(http.StatusNotAcceptable)

	// 1 sound
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id}
	httpexpect.New(t, sut.URL).
		POST("/sound/groups/").
		WithJSON(g1).
		Expect().
		Status(http.StatusNotAcceptable)

	// valid
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}
	httpexpect.New(t, sut.URL).
		POST("/sound/groups/").
		WithJSON(g1).
		Expect().
		Status(http.StatusCreated).
		JSON()
}

func TestUpdateGroup(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	s3 := NewSound()
	s3.Name = "s3"
	s3.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}

	_ = groupProvider.Save(&g1)

	body := g1

	// not found
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/groups/%s/", "foobar")).
		WithJSON(&body).
		Expect().
		Status(http.StatusNotFound)

	// bad name
	body.Name = ""
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/groups/%s/", g1.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusNotAcceptable)

	// bad sound id
	body = g1
	body.SoundIds = append(body.SoundIds, s3.Id)
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/groups/%s/", g1.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusNotAcceptable)

	// 1 sound
	body = g1
	body.SoundIds = []string{s1.Id}
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/groups/%s/", g1.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusNotAcceptable)

	// valid
	body = g1
	httpexpect.New(t, sut.URL).
		PATCH(fmt.Sprintf("/sound/groups/%s/", g1.Id)).
		WithJSON(&body).
		Expect().
		Status(http.StatusAccepted)

	assert.Equal(t, body.Name, groupProvider.Get(body.Id).Name)
	assert.Equal(t, body.SoundIds, groupProvider.Get(body.Id).SoundIds)
}

func TestDeleteGroup(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}

	_ = groupProvider.Save(&g1)

	// bad id
	httpexpect.New(t, sut.URL).
		DELETE(fmt.Sprintf("/sound/groups/%s/", "foobar")).
		Expect().
		Status(http.StatusNoContent)

	// valid
	httpexpect.New(t, sut.URL).
		DELETE(fmt.Sprintf("/sound/groups/%s/", g1.Id)).
		Expect().
		Status(http.StatusNoContent)
}

func TestPlayGroup(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = false

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}

	_ = groupProvider.Save(&g1)

	// bad id
	httpexpect.New(t, sut.URL).
		PUT(fmt.Sprintf("/sound/groups/%s/play/", "foobar")).
		Expect().
		Status(http.StatusNotFound)

	// valid
	httpexpect.New(t, sut.URL).
		PUT(fmt.Sprintf("/sound/groups/%s/play/", g1.Id)).
		Expect().
		Status(http.StatusAccepted)

	assert.Len(t, svc.playQueue.sounds, 2)
	assert.Equal(t, svc.playQueue.sounds[0].Id, s1.Id)
	assert.Equal(t, svc.playQueue.sounds[1].Id, s2.Id)
}

func TestSay(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	httpexpect.New(t, sut.URL).
		PUT("/sound/say/").
		WithJSON("hello world").
		Expect().
		Status(http.StatusAccepted)

	assert.Len(t, svc.playQueue.sounds, 1)

	content := bytes.NewBuffer([]byte{})
	err := soundProvider.ReadAudio(&svc.playQueue.sounds[0], content)
	if err != nil {
		t.Fail()
	}

	assert.Greater(t, len(content.Bytes()), 0)
}

func TestSearch(t *testing.T) {
	setup()

	sut := newServer()
	defer sut.Close()

	s1 := NewSound()
	s1.Name = "s1"
	s1.Hidden = false

	s2 := NewSound()
	s2.Name = "s2"
	s2.Hidden = true

	_ = soundProvider.Save(&s1)
	_ = soundProvider.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}

	_ = groupProvider.Save(&g1)

	// hidden sound
	httpexpect.New(t, sut.URL).
		GET("/sound/search/").
		WithQuery("q", "s2").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("sounds").
		ContainsKey("groups").
		ValueEqual("sounds", []Sound{})

	// no query
	httpexpect.New(t, sut.URL).
		GET("/sound/search/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("sounds").
		ContainsKey("groups").
		ValueEqual("sounds", []Sound{s1}).
		ValueEqual("groups", []Group{g1})

	// query
	httpexpect.New(t, sut.URL).
		GET("/sound/search/").
		WithQuery("q", "s1").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("sounds").
		ContainsKey("groups").
		ValueEqual("sounds", []Sound{s1}).
		ValueEqual("groups", []Group{})

}
