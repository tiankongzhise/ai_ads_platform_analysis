package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisStateStore struct {
	addr   string
	prefix string
}

func NewRedisStateStore(addr string) *RedisStateStore {
	return &RedisStateStore{addr: addr, prefix: "eduadcrm:oauth:state:"}
}

func (s *RedisStateStore) SaveState(state OAuthState) {
	payload, err := json.Marshal(state)
	if err != nil {
		return
	}
	ttl := time.Until(state.ExpiresAt)
	if ttl <= 0 {
		return
	}
	_, _ = s.command("SET", s.prefix+state.State, string(payload), "EX", strconv.Itoa(int(ttl.Seconds())))
}

func (s *RedisStateStore) ConsumeState(stateValue string) (OAuthState, bool) {
	payload, err := s.command("GETDEL", s.prefix+stateValue)
	if err != nil {
		return OAuthState{}, false
	}
	var state OAuthState
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return OAuthState{}, false
	}
	if time.Now().UTC().After(state.ExpiresAt) {
		return OAuthState{}, false
	}
	now := time.Now().UTC()
	state.UsedAt = &now
	return state, true
}

func (s *RedisStateStore) command(args ...string) (string, error) {
	conn, err := net.DialTimeout("tcp", s.addr, 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte(encodeRESP(args...))); err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)
	prefix, err := reader.ReadByte()
	if err != nil {
		return "", err
	}
	switch prefix {
	case '+':
		line, err := reader.ReadString('\n')
		return strings.TrimSpace(line), err
	case '$':
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		length, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return "", err
		}
		if length < 0 {
			return "", errors.New("redis nil")
		}
		buf := make([]byte, length+2)
		if _, err := reader.Read(buf); err != nil {
			return "", err
		}
		return string(buf[:length]), nil
	case ':':
		line, err := reader.ReadString('\n')
		return strings.TrimSpace(line), err
	case '-':
		line, _ := reader.ReadString('\n')
		return "", errors.New(strings.TrimSpace(line))
	default:
		return "", fmt.Errorf("unexpected redis response prefix %q", prefix)
	}
}

func encodeRESP(args ...string) string {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(args)))
	builder.WriteString("\r\n")
	for _, arg := range args {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(arg)))
		builder.WriteString("\r\n")
		builder.WriteString(arg)
		builder.WriteString("\r\n")
	}
	return builder.String()
}

type StateRepository struct {
	Repository
	stateStore StateStore
}

func NewStateRepository(base Repository, stateStore StateStore) *StateRepository {
	return &StateRepository{Repository: base, stateStore: stateStore}
}

func (r *StateRepository) SaveState(state OAuthState) {
	r.stateStore.SaveState(state)
}

func (r *StateRepository) ConsumeState(stateValue string) (OAuthState, bool) {
	return r.stateStore.ConsumeState(stateValue)
}

var _ Repository = (*StateRepository)(nil)
