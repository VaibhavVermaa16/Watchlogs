package helper

import "time"

func Tokenize(input string) []string {
	var tokens []string
	current := ""

	for _, char := range input {
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A' // Convert to lowercase
		}
		if char >= 'a' && char <= 'z' {
			current += string(char)
		} else {
			if current != "" {
				tokens = append(tokens, current)
				current = ""
			}
		}
	}
	if current != "" {
		tokens = append(tokens, current)
	}
	return tokens
}

func Intersect(a, b []int) []int {
	var i = 0
	var j = 0
	var result []int

	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			result = append(result, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return result
}

func ParseSince(since string) time.Time {
	if since == "" {
		return time.Now()
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		return time.Now()
	}
	
	return time.Now().Add(-duration)
}