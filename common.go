package parse // import "github.com/tdewolff/parse"

import (
	"encoding/base64"
	"errors"
	"net/url"
)

var ErrBadDataURI = errors.New("not a data URI")

func Number(b []byte) (n int, ok bool) {
	i := 0
	if i >= len(b) {
		return 0, false
	}
	if b[i] == '+' || b[i] == '-' {
		i++
		if i >= len(b) {
			return 0, false
		}
	}
	firstDigit := (b[i] >= '0' && b[i] <= '9')
	if firstDigit {
		i++
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
		}
	}
	if i < len(b) && b[i] == '.' {
		i++
		if i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
			for i < len(b) && b[i] >= '0' && b[i] <= '9' {
				i++
			}
		} else if firstDigit {
			// . could belong to the next token
			i--
			return i, true
		} else {
			return 0, false
		}
	} else if !firstDigit {
		return 0, false
	}
	iOld := i
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		i++
		if i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		if i >= len(b) || b[i] < '0' || b[i] > '9' {
			// e could belong to next token
			return iOld, true
		}
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
		}
	}
	return i, true
}

// DataURI splits the given URLToken and returns the mediatype, data and ok.
func DataURI(dataURI []byte) ([]byte, []byte, error) {
	if len(dataURI) > 5 && Equal(dataURI[:5], []byte("data:")) {
		dataURI = dataURI[5:]
		inBase64 := false
		mediatype := []byte{}
		i := 0
		for j, c := range dataURI {
			if c == '=' || c == ';' || c == ',' {
				if c != '=' && Equal(Trim(dataURI[i:j], IsWhitespace), []byte("base64")) {
					if len(mediatype) > 0 {
						mediatype = mediatype[:len(mediatype)-1]
					}
					inBase64 = true
					i = j
				} else if c != ',' {
					mediatype = append(append(mediatype, Trim(dataURI[i:j], IsWhitespace)...), c)
					i = j + 1
				} else {
					mediatype = append(mediatype, Trim(dataURI[i:j], IsWhitespace)...)
				}
				if c == ',' {
					if len(mediatype) == 0 || mediatype[0] == ';' {
						mediatype = []byte("text/plain")
					}
					data := dataURI[j+1:]
					if inBase64 {
						decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
						n, err := base64.StdEncoding.Decode(decoded, data)
						if err != nil {
							return []byte{}, []byte{}, err
						}
						data = decoded[:n]
					} else if unescaped, err := url.QueryUnescape(string(data)); err == nil {
						data = []byte(unescaped)
					}
					return mediatype, data, nil
				}
			}
		}
	}
	return []byte{}, []byte{}, ErrBadDataURI
}

func QuoteEntity(b []byte) (quote byte, n int, ok bool) {
	if len(b) < 5 || b[0] != '&' {
		return 0, 0, false
	}
	if b[1] == '#' {
		if b[2] == 'x' {
			i := 3
			for i < len(b) && b[i] == '0' {
				i++
			}
			if i+2 < len(b) && b[i] == '2' && b[i+2] == ';' {
				if b[i+1] == '2' {
					return '"', i + 3, true // &#x22;
				} else if b[i+1] == '7' {
					return '\'', i + 3, true // &#x27;
				}
			}
		} else {
			i := 2
			for i < len(b) && b[i] == '0' {
				i++
			}
			if i+2 < len(b) && b[i] == '3' && b[i+2] == ';' {
				if b[i+1] == '4' {
					return '"', i + 3, true // &#34;
				} else if b[i+1] == '9' {
					return '\'', i + 3, true // &#39;
				}
			}
		}
	} else if len(b) >= 6 && b[5] == ';' {
		if EqualCaseInsensitive(b[1:5], []byte{'q', 'u', 'o', 't'}) {
			return '"', 6, true // &quot;
		} else if EqualCaseInsensitive(b[1:5], []byte{'a', 'p', 'o', 's'}) {
			return '\'', 6, true // &apos;
		}
	}
	return 0, 0, false
}