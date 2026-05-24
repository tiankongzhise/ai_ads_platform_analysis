package store

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisBlacklist struct {
	addr   string
	prefix string
}

func NewRedisBlacklist(addr string) *RedisBlacklist {
	return &RedisBlacklist{
		addr:   addr,
		prefix: "eduadcrm:auth:blacklist:",
	}
}

func (r *RedisBlacklist) BlacklistAccessToken(tokenHash string, expiresAt time.Time) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return
	}
	_, _ = r.command("SET", r.prefix+tokenHash, "1", "EX", strconv.Itoa(int(ttl.Seconds())))
}

func (r *RedisBlacklist) IsBlacklisted(tokenHash string) bool {
	value, err := r.command("GET", r.prefix+tokenHash)
	return err == nil && value == "1"
}

func (r *RedisBlacklist) command(args ...string) (string, error) {
	conn, err := net.DialTimeout("tcp", r.addr, 2*time.Second)
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

type BlacklistRepository struct {
	Repository
	blacklist interface {
		BlacklistAccessToken(string, time.Time)
		IsBlacklisted(string) bool
	}
}

func NewBlacklistRepository(base Repository, blacklist interface {
	BlacklistAccessToken(string, time.Time)
	IsBlacklisted(string) bool
}) *BlacklistRepository {
	return &BlacklistRepository{Repository: base, blacklist: blacklist}
}

func (r *BlacklistRepository) BlacklistAccessToken(tokenHash string, expiresAt time.Time) {
	r.blacklist.BlacklistAccessToken(tokenHash, expiresAt)
}

func (r *BlacklistRepository) IsBlacklisted(tokenHash string) bool {
	return r.blacklist.IsBlacklisted(tokenHash)
}

var _ Repository = (*BlacklistRepository)(nil)
