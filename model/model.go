package model

import (
	"regexp"
	"time"
)

// Message for communication through a queue.
type Message = map[string]string

// uuidMatcher can be tested here: https://regex101.com/r/jh0MuE/latest
var uuidMatcher = regexp.MustCompile(`^[a-fA-F0-9]{8}(?:-?[a-fA-F0-9]{4}){3}-?[a-fA-F0-9]{12}$`)

type UUID string

func (u UUID) IsValid() bool {
	return uuidMatcher.MatchString(string(u))
}

func (u UUID) String() string {
	return string(u)
}

type Newsletter struct {
	ID      string
	Title   string
	Body    string
	Created time.Time
	Updated time.Time
}
