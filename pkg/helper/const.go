package helper

import "regexp"

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]{3,32}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	SECRET_KEY    = []byte("f0i09Z0jmu8pj6fkaGciYhDXIKRk1ozY")
)
