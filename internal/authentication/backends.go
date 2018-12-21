package authentication

import (
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Backend interface {
	TTL() time.Duration
	UserId(ipAddress string, token string) (string, error)
	NewToken(ipAddress string, userId string) (string, error)
	InvalidateToken(ipAddress string, token string)
}

func NewMemoryBackend(ttl time.Duration) MemoryBackend {
	var userIdMap map[string]map[string]string
	var timeoutMap map[string]map[string]time.Time

	return MemoryBackend{ttl, userIdMap, timeoutMap}
}

type MemoryBackend struct {
	ttl        time.Duration
	userIdMap  map[string]map[string]string
	timeoutMap map[string]map[string]time.Time
}

func (mb MemoryBackend) TTL() time.Duration {
	return mb.ttl
}

func (mb MemoryBackend) UserId(ipAddress string, token string) (string, error) {
	if userId, ok := mb.userIdMap[ipAddress][token]; ok {
		if mb.timeoutMap[ipAddress][token].Unix() > time.Now().Unix() {
			return userId, nil
		} else {
			mb.InvalidateToken(ipAddress, token)
		}
	}

	return "", errors.New("invalid ip token pair")
}

func (mb MemoryBackend) NewToken(ipAddress string, userId string) (string, error) {
	var token = strings.Replace(uuid.New().String(), "-", "", 4)

	mb.userIdMap[ipAddress][token] = userId
	mb.timeoutMap[ipAddress][token] = time.Now().Add(mb.ttl)

	return token, nil
}

func (mb MemoryBackend) InvalidateToken(ipAddress string, token string) {
	if ipMap, ok := mb.userIdMap[ipAddress]; ok {
		if len(ipMap) <= 1 {
			delete(mb.userIdMap, ipAddress)
			delete(mb.timeoutMap, ipAddress)
		} else {
			delete(mb.userIdMap[ipAddress], token)
			delete(mb.timeoutMap[ipAddress], token)
		}
	}
}
