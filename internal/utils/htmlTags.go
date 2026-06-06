package utils

import (
	"bytes"
)

func CheckTags(buffer []byte) int {
	bodyTag := []byte("</body>")
	htmlTag := []byte("</html>")

	if index := bytes.LastIndex(buffer, bodyTag); index != -1 {
		return index
	}

	if index := bytes.LastIndex(buffer, htmlTag); index != -1 {
		return index
	}

	return -1
}
