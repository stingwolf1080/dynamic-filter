package types

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"hash"
	"math"
	"strings"
)

type Hasher struct {
	HashName string
	Digest   func() hash.Hash
}

type OTP struct {
	secret  string // secret in base32 format
	expired int64
	digits  int     // number of integers in the OTP. Some apps expect this to be 6 digits, others support more.
	hasher  *Hasher // digest function to use in the HMAC (expected to be sha1)
}

func NewOTP(secret string, digits int, expired int64, hasher *Hasher) OTP {
	if hasher == nil {
		hasher = &Hasher{
			HashName: "sha1",
			Digest:   sha1.New,
		}
	}
	secret_base32 := base32.StdEncoding.EncodeToString([]byte(secret))
	return OTP{
		secret:  secret_base32,
		expired: expired,
		digits:  digits,
		hasher:  hasher,
	}
}

func (o *OTP) GenerateOTP() string {
	return o.generateOTP(o.expired)
}

/*
params

	input: the HMAC counter value to use as the OTP input. Usually either the counter, or the computed integer based on the Unix timestamp
*/
func (o *OTP) generateOTP(input int64) string {
	if input < 0 {
		panic("input must be positive integer")
	}
	hasher := hmac.New(o.hasher.Digest, o.byteSecret())
	hasher.Write(Itob(input))
	hmacHash := hasher.Sum(nil)

	offset := int(hmacHash[len(hmacHash)-1] & 0xf)
	code := ((int(hmacHash[offset]) & 0x7f) << 24) |
		((int(hmacHash[offset+1] & 0xff)) << 16) |
		((int(hmacHash[offset+2] & 0xff)) << 8) |
		(int(hmacHash[offset+3]) & 0xff)

	code = code % int(math.Pow10(o.digits))
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", o.digits), code)
}

func (o *OTP) VerifyOTP(otp string) bool {
	// fmt.Printf(otp, o.GenerateOTP())
	return otp == o.GenerateOTP()
}

func (o *OTP) byteSecret() []byte {
	missingPadding := len(o.secret) % 8
	if missingPadding != 0 {
		o.secret = o.secret + strings.Repeat("=", 8-missingPadding)
	}
	bytes, err := base32.StdEncoding.DecodeString(o.secret)
	if err != nil {
		panic("decode secret failed")
	}
	return bytes
}

const (
	OtpTypeTotp = "totp"
	OtpTypeHotp = "hotp"
)

// integer to byte array
func Itob(integer int64) []byte {
	byteArr := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		byteArr[i] = byte(integer & 0xff)
		integer = integer >> 8
	}
	return byteArr
}

// A non-panic way of seeing weather or not a given secret is valid
func IsSecretValid(secret string) bool {
	missingPadding := len(secret) % 8
	if missingPadding != 0 {
		secret = secret + strings.Repeat("=", 8-missingPadding)
	}
	_, err := base32.StdEncoding.DecodeString(secret)
	return err == nil
}
