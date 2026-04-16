package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

// Parse verifies if it has \r\n\r\n and processes just the first, the done come only when there's only \r\n
func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	dataStr := string(data)
	//log.Printf("%s\n", data)
	done = false

	if strings.HasPrefix(dataStr, "\r\n") {
		return 2, true, nil
	}

	index := strings.Index(dataStr, "\r\n")

	if index == -1 {
		return 0, false, nil
	}

	dataStr = dataStr[:index]
	index = strings.Index(dataStr, ":")

	if index == -1 {
		return 0, false, fmt.Errorf("malformed header: missing colon")
	}

	if index == 0 {
		return 0, false, fmt.Errorf("malformed headers")
	}

	if containsUnallowedCharacters(dataStr[:index]) {
		return 0, false, fmt.Errorf("malformed headers: unallowed characters")
	}

	value := dataStr[index+1:]

	value = strings.TrimSpace(value)

	key := strings.ToLower(dataStr[:index])

	if key == "" || strings.Contains(dataStr[:index], " ") {
		return 0, false, fmt.Errorf("malformed headers: invalid key format")
	}

	//fmt.Println(key)
	//fmt.Println(value)
	if _, ok := h[key]; !ok {
		h[key] = value
	} else {
		h[key] += ", " + value
	}
	//fmt.Println(h[key])

	return len(dataStr) + 2, done, nil
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func containsUnallowedCharacters(s string) bool {
	for _, r := range s {
		if !allowed(r) {
			return true
		}
	}
	return false
}

func allowed(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '!' || r == '#' || r == '$' || r == '%' || r == '&' || r == '\'' ||
		r == '*' || r == '+' || r == '-' || r == '.' || r == '^' ||
		r == '_' || r == '`' || r == '|' || r == '~'
}

func NewHeaders() Headers {
	return make(map[string]string)
}
