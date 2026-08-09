package sandbox

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type tokenIssuer struct {
	secret []byte
	now    func() time.Time
}

func newTokenIssuer() (*tokenIssuer, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate local auth key: %w", err)
	}
	return &tokenIssuer{secret: secret, now: time.Now}, nil
}

func (i *tokenIssuer) issue(subject, kind string, ttl time.Duration) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claims, err := json.Marshal(map[string]any{
		"sub": subject, "kind": kind, "iat": i.now().Unix(), "exp": i.now().Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	unsigned := encodeSegment(header) + "." + encodeSegment(claims)
	signature := hmac.New(sha256.New, i.secret)
	_, _ = signature.Write([]byte(unsigned))
	return unsigned + "." + encodeSegment(signature.Sum(nil)), nil
}

func (i *tokenIssuer) validate(token, kind string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("malformed token")
	}
	signature := hmac.New(sha256.New, i.secret)
	_, _ = signature.Write([]byte(parts[0] + "." + parts[1]))
	provided, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature.Sum(nil), provided) {
		return errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid token payload")
	}
	var claims struct {
		Kind string `json:"kind"`
		Exp  int64  `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return errors.New("invalid token claims")
	}
	if claims.Kind != kind {
		return errors.New("unexpected token kind")
	}
	if i.now().Unix() >= claims.Exp {
		return errors.New("token expired")
	}
	return nil
}

func encodeSegment(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
