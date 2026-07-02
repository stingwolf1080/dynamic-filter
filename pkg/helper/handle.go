package helper

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/jinzhu/inflection"
)

func IsUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func IsLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func ToUpper(c byte) byte {
	return c - 32
}

func ToLower(c byte) byte {
	return c + 32
}

func ConvertName(s string) string {
	trans := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(trans, s)
	s = strings.ReplaceAll(s, "Đ", "D")

	s = regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "_")
	return strings.ToUpper(s)
}

// Underscore converts "CamelCasedString" to "camel_cased_string".
func Underscore(s string) string {
	r := make([]byte, 0, len(s)+5)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			if i > 0 && i+1 < len(s) && (IsLower(s[i-1]) || IsLower(s[i+1])) {
				r = append(r, '_', ToLower(c))
			} else {
				r = append(r, ToLower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

// Un Underscore converts "camel_cased_string" to "CamelCasedString".
func ReverseUnderscore(s string) string {
	r := make([]byte, 0, len(s))
	upperNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			if IsLower(c) {
				c = ToUpper(c)
			}
			upperNext = false
		}
		r = append(r, c)
	}
	return string(r)
}

func UnderscoreToSpace(s string) string {
	r := make([]byte, 0, len(s)+5)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			if i > 0 && i+1 < len(s) && (IsLower(s[i-1]) || IsLower(s[i+1])) {
				r = append(r, ' ', ToLower(c))
			} else {
				r = append(r, ToLower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func GetModelName(myvar interface{}) string {
	name := getType(myvar)
	modeltbl := UnderscoreToSpace(name)
	return modeltbl
}

func GetModelTable(myvar interface{}) string {
	name := getType(myvar)
	modeltbl := Underscore(name)
	tblname := inflection.Plural(modeltbl)
	return tblname
}

func GetModelTableGeneric[T any]() string {
	name := GetNameModel[T]()
	modeltbl := Underscore(name)
	return inflection.Plural(modeltbl)
}

func GetModelTableFromString(name string) string {
	modeltbl := Underscore(name)
	tblname := inflection.Plural(modeltbl)
	return tblname
}

func GetStructName(tbl_name string) string {
	s := strings.ReplaceAll(tbl_name, "leaves", "leave")
	tblname := inflection.Singular(s)
	st_name := ReverseUnderscore(tblname)
	return st_name
}

func GetNameModel[T any]() string {
	var t T
	typ := reflect.TypeOf(t)
	if typ == nil {
		return ""
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return strings.ToLower(
		strings.ReplaceAll(typ.PkgPath(), "/", "_") +
			"_" +
			typ.Name(),
	)
}

func GetRoleName(model string) string {
	modeltbl := Underscore(model)
	return modeltbl
}

func GetRoleNameFromString(model string) string {
	modeltbl := UnderscoreToSpace(model)
	return modeltbl
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func CheckStringInArrayString(rep []string, s string) bool {
	for _, v := range rep {
		if v == s {
			return true
		}
	}
	return false
}

func TypeName[T any]() string {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func IsEmail(s string) bool {
	return emailRegex.MatchString(s)
}

func IsUsername(s string) bool {
	return usernameRegex.MatchString(s)
}
