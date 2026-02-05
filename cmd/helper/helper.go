package helper

func Tokenize(input string) []string {
	var tokens []string
	current := ""

	for _, char := range input {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
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
