package helper

func Tokenize(input string) []string {
	var tokens []string
	current :=""

	for _, char := range input {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			current += string(char)
		}else {
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