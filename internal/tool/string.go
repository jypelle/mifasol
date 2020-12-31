package tool

func CharacterTruncate(source string, characterLength int) string {
	runes := []rune(source)
	if len(runes) > characterLength {
		if characterLength > 3 {
			return string(runes[:characterLength-3]) + "..."
		} else {
			return "..."[:characterLength]
		}
	} else {
		return string(runes[0:len(runes)])
	}
}

func ByteTruncate(source string, byteLength int) string {
	runes := []rune(source)
	if len(runes) > byteLength {
		if byteLength > 3 {
			return string(runes[:byteLength-3]) + "..."
		} else {
			return "..."[:byteLength]
		}
	} else {
		return string(runes[0:len(runes)])
	}
}
