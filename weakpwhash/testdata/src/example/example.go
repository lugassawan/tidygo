package example

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	oldpbkdf2 "golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const fixedSalt = "static-salt"

func stdPBKDF2(pw string, salt []byte, iter int) {
	h := sha256.New

	_, _ = pbkdf2.Key(h, pw, nil, 600000, 32)               // want `pbkdf2.Key uses a constant salt`
	_, _ = pbkdf2.Key(h, pw, []byte("salt"), 600000, 32)    // want `pbkdf2.Key uses a constant salt`
	_, _ = pbkdf2.Key(h, pw, []byte{1, 2, 3}, 600000, 32)   // want `pbkdf2.Key uses a constant salt`
	_, _ = pbkdf2.Key(h, pw, []byte(fixedSalt), 600000, 32) // want `pbkdf2.Key uses a constant salt`
	_, _ = pbkdf2.Key(
		h,
		pw,
		salt,
		1000,
		32,
	) // want `pbkdf2.Key iteration count 1000 is below the minimum 100000`
	_, _ = pbkdf2.Key(
		h,
		pw,
		salt,
		99999,
		32,
	) // want `pbkdf2.Key iteration count 99999 is below the minimum 100000`
	_, _ = pbkdf2.Key(
		h,
		pw,
		nil,
		1000,
		32,
	) // want `pbkdf2.Key uses a constant salt` `pbkdf2.Key iteration count 1000 is below`
	_, _ = pbkdf2.Key(h, pw, salt, 100000, 32)                         // at the floor
	_, _ = pbkdf2.Key(h, pw, salt, iter, 32)                           // non-constant cost
	_, _ = pbkdf2.Key(h, pw, append([]byte("a"), salt...), 600000, 32) // variable content
}

func oldPBKDF2(pw, salt []byte, iter int) {
	h := sha256.New

	_ = oldpbkdf2.Key(pw, nil, 600000, 32, h)         // want `pbkdf2.Key uses a constant salt`
	_ = oldpbkdf2.Key(pw, []byte("s"), 600000, 32, h) // want `pbkdf2.Key uses a constant salt`
	_ = oldpbkdf2.Key(pw, salt, 4096, 32, h)          // want `pbkdf2.Key iteration count 4096 is below the minimum`
	_ = oldpbkdf2.Key(pw, salt, 100000, 32, h)
	_ = oldpbkdf2.Key(pw, salt, iter, 32, h)
}

func scryptKey(pw, salt []byte, n int) {
	_, _ = scrypt.Key(pw, []byte("salt"), 32768, 8, 1, 32) // want `scrypt.Key uses a constant salt`
	_, _ = scrypt.Key(pw, salt, 16384, 8, 1, 32)           // want `scrypt.Key N 16384 is below the minimum 32768`
	_, _ = scrypt.Key(pw, salt, 32768, 8, 1, 32)
	_, _ = scrypt.Key(pw, salt, n, 8, 1, 32)
}

func argon2Keys(pw, salt []byte, mem uint32) {
	_ = argon2.Key(pw, nil, 1, 65536, 4, 32)           // want `argon2.Key uses a constant salt`
	_ = argon2.IDKey(pw, []byte("s"), 1, 65536, 4, 32) // want `argon2.IDKey uses a constant salt`
	_ = argon2.Key(pw, salt, 1, 4096, 4, 32)           // want `argon2.Key memory 4096 is below the minimum 19456`
	_ = argon2.IDKey(pw, salt, 1, 19455, 4, 32)        // want `argon2.IDKey memory 19455 is below the minimum 19456`
	_ = argon2.IDKey(pw, salt, 1, 19456, 4, 32)
	_ = argon2.IDKey(pw, salt, 1, mem, 4, 32)
}

func bcryptCost(pw []byte, cost int) {
	_, _ = bcrypt.GenerateFromPassword(pw, 4)  // want `bcrypt.GenerateFromPassword cost 4 is below the minimum 10`
	_, _ = bcrypt.GenerateFromPassword(pw, 9)  // want `bcrypt.GenerateFromPassword cost 9 is below the minimum 10`
	_, _ = bcrypt.GenerateFromPassword(pw, 10) // at the floor
	_, _ = bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
	_, _ = bcrypt.GenerateFromPassword(pw, 0) // bcrypt substitutes DefaultCost
	_, _ = bcrypt.GenerateFromPassword(pw, 3) // bcrypt substitutes DefaultCost
	_, _ = bcrypt.GenerateFromPassword(pw, cost)
}

func randomSalt(pw string) {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	_, _ = pbkdf2.Key(sha256.New, pw, salt, 600000, 32)
}
