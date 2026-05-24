package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ValueType string

const (
	TypeString    ValueType = "string"
	TypeDuration  ValueType = "duration"
	TypeBool      ValueType = "bool"
	TypeInt       ValueType = "int"
	TypeSecretRef ValueType = "secret_ref"
)

type Definition struct {
	Key          string    `json:"key"`
	Type         ValueType `json:"type"`
	DefaultValue string    `json:"default_value"`
	Editable     bool      `json:"editable"`
	Sensitive    bool      `json:"sensitive"`
	Description  string    `json:"description"`
}

type Value struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Source    string    `json:"source"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PublicConfig struct {
	Definitions []Definition `json:"definitions"`
	Values      []Value      `json:"values"`
}

type AuthConfig struct {
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	RefreshTokenRotation bool
	LogoutBlacklist      bool
	Issuer               string
	JWTSecret            string
	PasswordIterations   int
}

func ValidateValue(def Definition, value string) error {
	switch def.Type {
	case TypeString:
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", def.Key)
		}
	case TypeDuration:
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("%s must be a duration: %w", def.Key, err)
		}
	case TypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("%s must be a boolean: %w", def.Key, err)
		}
	case TypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("%s must be an integer: %w", def.Key, err)
		}
	case TypeSecretRef:
		if !strings.Contains(value, "://") {
			return fmt.Errorf("%s must be a secret reference URI", def.Key)
		}
	default:
		return fmt.Errorf("unsupported config type %q", def.Type)
	}
	return nil
}
