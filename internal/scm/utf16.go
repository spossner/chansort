package scm

// Minimal UTF-16BE zero-terminated decoder (BMP only; enough for channel names)
func decodeUTF16BEZeroTerm(b []byte) string {
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		if b[i] == 0 && b[i+1] == 0 {
			break
		}
		cp := uint16(b[i])<<8 | uint16(b[i+1])
		runes = append(runes, rune(cp))
	}
	return string(runes)
}
