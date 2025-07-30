package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	dataStr := string(data)

	if !strings.Contains(dataStr, "\r\n") {
		return 0, false, nil
	}

	if strings.HasPrefix(dataStr, "\r\n") {
		return 1, true, nil
	}

	index := strings.Index(dataStr, "\r\n")
	dataStr = dataStr[:index]
	index = strings.Index(dataStr, ":")

	if index == 0 {
		return 0, false, fmt.Errorf("malformed headers")
	}

	if strings.HasSuffix(dataStr[:index], " ") {
		return 0, false, fmt.Errorf("malformed headers")
	}

	if containsUnallowedCharacters(dataStr[:index]) {
		return 0, false, fmt.Errorf("malformed headers: unallowed characters")
	}

	value := dataStr[index+1:]

	value = strings.TrimSpace(value)

	//value = strings.ToLower(value)

	key := strings.ToLower(dataStr[:index])
	fmt.Println(key)
	fmt.Println(value)
	if _, ok := h[key]; !ok {
		h[key] = value
	} else {
		h[key] += ", " + value
	}
	fmt.Println(h[key])

	return len(dataStr) + 2, false, nil
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
