package api

import (
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type AuthenticationBackend interface {
	TTL() time.Duration
	UserId(ipAddress string, token string) (string, error)
	NewToken(ipAddress string, userId string) (string, error)
	InvalidateToken(ipAddress string, token string)
}

type AuthenticationMemoryBackend struct {
	ttl        time.Duration
	userIdMap  map[string]map[string]string
	timeoutMap map[string]map[string]time.Time
}

func NewAuthenticationMemoryBackend(ttl time.Duration) AuthenticationMemoryBackend {
	var userIdMap map[string]map[string]string
	var timeoutMap map[string]map[string]time.Time

	return AuthenticationMemoryBackend{ttl, userIdMap, timeoutMap}
}

func (mb AuthenticationMemoryBackend) TTL() time.Duration {
	return mb.ttl
}

func (mb AuthenticationMemoryBackend) UserId(ipAddress string, token string) (string, error) {
	if userId, ok := mb.userIdMap[ipAddress][token]; ok {
		if mb.timeoutMap[ipAddress][token].Unix() > time.Now().Unix() {
			return userId, nil
		} else {
			mb.InvalidateToken(ipAddress, token)
		}
	}

	return "", errors.New("invalid ip token pair")
}

func (mb AuthenticationMemoryBackend) NewToken(ipAddress string, userId string) (string, error) {
	var token = strings.Replace(uuid.New().String(), "-", "", 4)

	mb.userIdMap[ipAddress][token] = userId
	mb.timeoutMap[ipAddress][token] = time.Now().Add(mb.ttl)

	return token, nil
}

func (mb AuthenticationMemoryBackend) InvalidateToken(ipAddress string, token string) {
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

type AuthenticationNoopBackend struct {
	ttl time.Duration
}

func NewAuthenticationNoopBackend(ttl time.Duration) *AuthenticationNoopBackend {
	return &AuthenticationNoopBackend{ttl: ttl}
}

func (b AuthenticationNoopBackend) TTL() time.Duration {
	return b.ttl
}

func (b AuthenticationNoopBackend) UserId(ipAddress string, token string) (string, error) {
	return "", nil
}

func (b AuthenticationNoopBackend) NewToken(ipAddress string, userId string) (string, error) {
	return "", nil
}

func (b AuthenticationNoopBackend) InvalidateToken(ipAddress string, token string) {
	return
}
