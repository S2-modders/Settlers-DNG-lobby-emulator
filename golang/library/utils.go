package library

func CStr2Str(input []byte) string {
	return string(input[:clen(input)])
}

func clen(n []byte) int {
    for i := 0; i < len(n); i++ {
        if n[i] == 0 {
            return i
        }
    }
    return len(n)
}