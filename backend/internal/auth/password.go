// Argon2id password hashing in PHC string format. Parameters are recorded in
// the hash itself so VerifyPassword can signal when a stored hash should be
// rehashed on next successful login.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 64 * 1024 // KiB
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

var ErrMalformedHash = errors.New("malformed password hash")

// HashPassword returns a PHC-format Argon2id hash:
// $argon2id$v=19$m=65536,t=2,p=2$<b64 salt>$<b64 key>
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword reports whether password matches the PHC hash and whether the
// stored parameters differ from the current policy (needsRehash).
func VerifyPassword(password, encoded string) (match, needsRehash bool, err error) {
	var (
		version          int
		memory, timeCost uint32
		threads          uint8
		saltB64, keyB64  string
	)
	parts := strings.Split(encoded, "$")
	// ["", "argon2id", "v=19", "m=...,t=...,p=...", salt, key]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, false, ErrMalformedHash
	}
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, false, ErrMalformedHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false, false, ErrMalformedHash
	}
	saltB64, keyB64 = parts[4], parts[5]
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, false, ErrMalformedHash
	}
	wantKey, err := base64.RawStdEncoding.DecodeString(keyB64)
	if err != nil {
		return false, false, ErrMalformedHash
	}

	got := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(wantKey)))
	match = subtle.ConstantTimeCompare(got, wantKey) == 1
	needsRehash = memory != argonMemory || timeCost != argonTime || threads != argonThreads ||
		uint32(len(wantKey)) != argonKeyLen
	return match, needsRehash, nil
}
