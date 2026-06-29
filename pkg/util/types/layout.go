package types

import "time"

const (
	rfc3339           = "2006-01-02"
	rfc822            = "2006-01-02 15:04"
	iso8601           = "2006-01-02 15:04:05"
	kitchen           = "15:04"
	NumStr            = 12
	letterBytes       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits     = 6                    // 6 bits to represent a letter index
	letterIdxMask     = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax      = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	LayoutISO         = "2006-01-02"
	LayoutISODateTime = "2006-01-02 15:04:05"
	LayoutUS          = "January 2, 2006"
	LayoutRFC3339     = time.RFC3339
	LayoutTime        = "15:04:05"
	MaxSize           = 5 << 20
	DefaultLanguage   = "english"
)

type SafeString string
type SafeInt string
