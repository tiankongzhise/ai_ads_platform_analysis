package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Loader interface {
	Load() (map[string]string, error)
}

type ChangeLog struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	Action    string    `json:"action"`
	Version   int       `json:"version"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	mu          sync.RWMutex
	loader      Loader
	definitions map[string]Definition
	values      map[string]Value
	overrides   map[string]Value
	logs        []ChangeLog
}

func NewManager(loader Loader) (*Manager, error) {
	manager := &Manager{
		loader:      loader,
		definitions: defaultDefinitions(),
		values:      map[string]Value{},
		overrides:   map[string]Value{},
	}
	if err := manager.Reload(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) Reload() error {
	values, err := m.loader.Load()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, def := range m.definitions {
		raw := values[key]
		if raw == "" {
			raw = def.DefaultValue
		}
		if err := ValidateValue(def, raw); err != nil {
			return err
		}
		if override, ok := m.overrides[key]; ok {
			m.values[key] = override
			continue
		}
		current := m.values[key]
		version := current.Version
		if version == 0 {
			version = 1
		}
		m.values[key] = Value{
			Key:       key,
			Value:     raw,
			Source:    "file",
			Version:   version,
			UpdatedAt: time.Now().UTC(),
		}
	}
	return nil
}

func (m *Manager) Public() PublicConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	definitions := make([]Definition, 0, len(m.definitions))
	values := make([]Value, 0, len(m.values))
	for _, def := range m.definitions {
		definitions = append(definitions, def)
	}
	for _, value := range m.values {
		if def, ok := m.definitions[value.Key]; ok && def.Sensitive {
			value.Value = "******"
		}
		values = append(values, value)
	}
	return PublicConfig{Definitions: definitions, Values: values}
}

func (m *Manager) Update(key string, value string, changedBy string) (Value, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	def, ok := m.definitions[key]
	if !ok {
		return Value{}, fmt.Errorf("unknown config key %s", key)
	}
	if !def.Editable {
		return Value{}, fmt.Errorf("config key %s is not editable", key)
	}
	if err := ValidateValue(def, value); err != nil {
		return Value{}, err
	}
	old := m.values[key]
	next := Value{
		Key:       key,
		Value:     value,
		Source:    "database",
		Version:   old.Version + 1,
		UpdatedAt: time.Now().UTC(),
	}
	m.values[key] = next
	m.overrides[key] = next
	m.logs = append(m.logs, ChangeLog{
		ID:        newID(),
		Key:       key,
		OldValue:  old.Value,
		NewValue:  value,
		Action:    "publish",
		Version:   next.Version,
		ChangedBy: changedBy,
		CreatedAt: time.Now().UTC(),
	})
	return next, nil
}

func (m *Manager) Rollback(key string, changedBy string) (Value, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	def, ok := m.definitions[key]
	if !ok {
		return Value{}, fmt.Errorf("unknown config key %s", key)
	}
	if !def.Editable {
		return Value{}, fmt.Errorf("config key %s is not editable", key)
	}
	current := m.values[key]
	rollbackValue := def.DefaultValue
	for i := len(m.logs) - 1; i >= 0; i-- {
		log := m.logs[i]
		if log.Key == key && log.OldValue != "" {
			rollbackValue = log.OldValue
			break
		}
	}
	if err := ValidateValue(def, rollbackValue); err != nil {
		return Value{}, err
	}
	next := Value{
		Key:       key,
		Value:     rollbackValue,
		Source:    "database",
		Version:   current.Version + 1,
		UpdatedAt: time.Now().UTC(),
	}
	m.values[key] = next
	m.overrides[key] = next
	m.logs = append(m.logs, ChangeLog{
		ID:        newID(),
		Key:       key,
		OldValue:  current.Value,
		NewValue:  rollbackValue,
		Action:    "rollback",
		Version:   next.Version,
		ChangedBy: changedBy,
		CreatedAt: time.Now().UTC(),
	})
	return next, nil
}

func (m *Manager) Logs() []ChangeLog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	logs := make([]ChangeLog, len(m.logs))
	copy(logs, m.logs)
	return logs
}

func (m *Manager) GetString(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[key].Value
}

func (m *Manager) Auth() (AuthConfig, error) {
	accessTTL, err := time.ParseDuration(m.GetString("auth.access_token_ttl"))
	if err != nil {
		return AuthConfig{}, err
	}
	refreshTTL, err := time.ParseDuration(m.GetString("auth.refresh_token_ttl"))
	if err != nil {
		return AuthConfig{}, err
	}
	rotation, err := strconv.ParseBool(m.GetString("auth.refresh_token_rotation"))
	if err != nil {
		return AuthConfig{}, err
	}
	blacklist, err := strconv.ParseBool(m.GetString("auth.logout_blacklist_enabled"))
	if err != nil {
		return AuthConfig{}, err
	}
	iterations, err := strconv.Atoi(m.GetString("auth.password_iterations"))
	if err != nil {
		return AuthConfig{}, err
	}
	return AuthConfig{
		AccessTokenTTL:       accessTTL,
		RefreshTokenTTL:      refreshTTL,
		RefreshTokenRotation: rotation,
		LogoutBlacklist:      blacklist,
		Issuer:               m.GetString("auth.issuer"),
		JWTSecret:            m.GetString("auth.jwt_secret"),
		PasswordIterations:   iterations,
	}, nil
}

func defaultDefinitions() map[string]Definition {
	defs := []Definition{
		{Key: "auth.access_token_ttl", Type: TypeDuration, DefaultValue: "24h", Editable: true, Description: "访问令牌有效期，默认 1 天。"},
		{Key: "auth.refresh_token_ttl", Type: TypeDuration, DefaultValue: "720h", Editable: true, Description: "刷新令牌有效期，默认 30 天。"},
		{Key: "auth.refresh_token_rotation", Type: TypeBool, DefaultValue: "true", Editable: true, Description: "刷新令牌时是否轮换 refresh token。"},
		{Key: "auth.logout_blacklist_enabled", Type: TypeBool, DefaultValue: "true", Editable: true, Description: "登出时是否将 access token 写入黑名单。"},
		{Key: "auth.issuer", Type: TypeString, DefaultValue: "eduadcrm", Editable: false, Description: "JWT 签发方。"},
		{Key: "auth.jwt_secret", Type: TypeString, DefaultValue: "dev-only-change-me", Editable: false, Sensitive: true, Description: "JWT HMAC 密钥，生产环境必须通过环境变量或密钥系统覆盖。"},
		{Key: "auth.password_iterations", Type: TypeInt, DefaultValue: "210000", Editable: false, Description: "PBKDF2 密码哈希迭代次数。"},
		{Key: "config.cache_refresh_interval", Type: TypeDuration, DefaultValue: "30s", Editable: true, Description: "服务内配置缓存刷新间隔。"},
	}
	out := map[string]Definition{}
	for _, def := range defs {
		out[def.Key] = def
	}
	return out
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
