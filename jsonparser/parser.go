package jsonparser

import (
	"bytes"
	"errors"
	"unicode/utf8"
)

// 基于jsonparser扩展实现，增加了注释解析 https://github.com/buger/jsonparser
// Errors
var (
	KeyPathNotFoundError       = errors.New("Key path not found")
	UnknownValueTypeError      = errors.New("Unknown value type")
	MalformedJsonError         = errors.New("Malformed JSON error")
	MalformedStringError       = errors.New("Value is string, but can't find closing '\"' symbol")
	MalformedArrayError        = errors.New("Value is array, but can't find closing ']' symbol")
	MalformedObjectError       = errors.New("Value looks like object, but can't find closing '}' symbol")
	MalformedValueError        = errors.New("Value looks like Number/Boolean/None, but can't find its end: ',' or '}' symbol")
	OverflowIntegerError       = errors.New("Value is number, but overflowed while parsing")
	MalformedStringEscapeError = errors.New("Encountered an invalid escape sequence in a string")
	NullValueError             = errors.New("Value is null")
)

var (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
)

// backslashCharEscapeTable: when '\X' is found for some byte X, it is to be replaced with backslashCharEscapeTable[X]
var backslashCharEscapeTable = [...]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

const unescapeStackBufSize = 64
const supplementalPlanesOffset = 0x10000
const highSurrogateOffset = 0xD800
const lowSurrogateOffset = 0xDC00
const basicMultilingualPlaneReservedOffset = 0xDFFF
const basicMultilingualPlaneOffset = 0xFFFF
const badHex = -1

// Data types available in valid JSON data.
type ValueType int

const (
	NotExist = ValueType(iota)
	String
	Number
	Object
	Array
	Boolean
	Null
	Unknown
)

func Get(data []byte) (value []byte, dataType ValueType, offset int, err error) {
	a, b, _, d, e := internalGet(data)
	return a, b, d, e
}

func internalGet(data []byte) (value []byte, dataType ValueType, offset, endOffset int, err error) {
	// Go to closest value
	nO := nextToken(data[offset:])
	if nO == -1 {
		return nil, NotExist, offset, -1, MalformedJsonError
	}

	offset += nO
	value, dataType, endOffset, err = getType(data, offset)
	if err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}

func getType(data []byte, offset int) ([]byte, ValueType, int, error) {
	var dataType ValueType
	endOffset := offset

	// if string value
	if data[offset] == '"' {
		dataType = String
		if idx, _ := stringEnd(data[offset+1:]); idx != -1 {
			endOffset += idx + 1
		} else {
			return nil, dataType, offset, MalformedStringError
		}
	} else if data[offset] == '[' { // if array value
		dataType = Array
		// break label, for stopping nested loops
		endOffset = blockEnd(data[offset:], '[', ']')

		if endOffset == -1 {
			return nil, dataType, offset, MalformedArrayError
		}

		endOffset += offset
	} else if data[offset] == '{' { // if object value
		dataType = Object
		// break label, for stopping nested loops
		endOffset = blockEnd(data[offset:], '{', '}')

		if endOffset == -1 {
			return nil, dataType, offset, MalformedObjectError
		}

		endOffset += offset
	} else {
		// Number, Boolean or None
		end := tokenEnd(data[endOffset:])

		if end == -1 {
			return nil, dataType, offset, MalformedValueError
		}

		value := data[offset : endOffset+end]

		switch data[offset] {
		case 't', 'f': // true or false
			if bytes.Equal(value, trueLiteral) || bytes.Equal(value, falseLiteral) {
				dataType = Boolean
			} else {
				return nil, Unknown, offset, UnknownValueTypeError
			}
		case 'u', 'n': // undefined or null
			if bytes.Equal(value, nullLiteral) {
				dataType = Null
			} else {
				return nil, Unknown, offset, UnknownValueTypeError
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			dataType = Number
		default:
			return nil, Unknown, offset, UnknownValueTypeError
		}

		endOffset += end
	}
	return data[offset:endOffset], dataType, endOffset, nil
}

// Tries to find the end of string
// Support if string contains escaped quote symbols.
func stringEnd(data []byte) (int, bool) {
	escaped := false
	for i, c := range data {
		if c == '"' {
			if !escaped {
				return i + 1, false
			} else {
				j := i - 1
				for {
					if j < 0 || data[j] != '\\' {
						return i + 1, true // even number of backslashes
					}
					j--
					if j < 0 || data[j] != '\\' {
						break // odd number of backslashes
					}
					j--

				}
			}
		} else if c == '\\' {
			escaped = true
		}
	}

	return -1, escaped
}

// Find end of the data structure, array or object.
// For array openSym and closeSym will be '[' and ']', for object '{' and '}'
func blockEnd(data []byte, openSym byte, closeSym byte) int {
	level := 0
	i := 0
	ln := len(data)

	for i < ln {
		switch data[i] {
		case '"': // If inside string, skip it
			se, _ := stringEnd(data[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym: // If open symbol, increase level
			level++
		case closeSym: // If close symbol, increase level
			level--

			// If we have returned to the original level, we're done
			if level == 0 {
				return i + 1
			}
		}
		i++
	}

	return -1
}

func tokenEnd(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}

	return len(data)
}

// Find position of next character which is not whitespace
func nextToken(data []byte) int {
	for i, c := range data {
		switch c {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}

// 对象迭代，如果flag为false，则停止迭代
func ObjectEach(data []byte, callback func(key []byte, value []byte, dataType ValueType, offset int, comment []byte) (bool, error)) error {
	offset := 0

	// Validate and skip past opening brace
	if off := nextToken(data[offset:]); off == -1 {
		return MalformedObjectError
	} else if offset += off; data[offset] != '{' {
		return MalformedObjectError
	} else {
		offset++
	}

	// Skip to the first token inside the object, or stop if we find the ending brace
	if off := nextToken(data[offset:]); off == -1 {
		return MalformedJsonError
	} else if offset += off; data[offset] == '}' {
		return nil
	}
	// 对象属性的注释，如果有多行，直接合并
	comment := make([]byte, 0)
	// Loop pre-condition: data[offset] points to what should be either the next entry's key, or the closing brace (if it's anything else, the JSON is malformed)
	for offset < len(data) {
		// Step 1: find the next key
		var key []byte

		// Check what the the next token is: start of string, end of object, or something else (error)
		switch data[offset] {
		case '"':
			offset++ // accept as string and skip opening quote
		case '}':
			return nil // we found the end of the object; stop and return success
		case '/':
			if data[offset] == '/' {
				end := commentEnd(data[offset:])
				if end != -1 {
					if len(comment) > 0 {
						comment = append(comment, '\n')
					}
					comment = append(comment, data[offset:offset+end]...)
					offset = offset + end
					if off := nextToken(data[offset:]); off == -1 {
						return MalformedObjectError
					} else {
						offset += off
					}
				} else {
					return MalformedObjectError
				}
				// 必须先匹配"，所以继续下次循环
				continue
			}
		default:
			return MalformedObjectError
		}

		// Find the end of the key string
		var keyEscaped bool
		if off, esc := stringEnd(data[offset:]); off == -1 {
			return MalformedJsonError
		} else {
			key, keyEscaped = data[offset:offset+off-1], esc
			offset += off
		}

		// Unescape the string if needed
		if keyEscaped {
			var stackbuf [unescapeStackBufSize]byte // stack-allocated array for allocation-free unescaping of small strings
			if keyUnescaped, err := Unescape(key, stackbuf[:]); err != nil {
				return MalformedStringEscapeError
			} else {
				key = keyUnescaped
			}
		}

		// Step 2: skip the colon
		if off := nextToken(data[offset:]); off == -1 {
			return MalformedJsonError
		} else if offset += off; data[offset] != ':' {
			return MalformedJsonError
		} else {
			offset++
		}

		// Step 3: find the associated value, then invoke the callback
		value, valueType, off, err := Get(data[offset:])
		if err != nil {
			return err
		} else {
			offset += off

		}

		// Step 4: skip over the next comma to the following token, or stop if we hit the ending brace
		if off = nextToken(data[offset:]); off == -1 {
			return MalformedArrayError
		} else {
			offset += off
			endFlag := false
			switch data[offset] {
			case ',':
				// 判断后面是否有注释
				offset++
				if data[offset+1] == '/' {
					offset++
					end := commentEnd(data[offset:])
					if end != -1 {
						if len(comment) > 0 {
							comment = append(comment, '\n')
						}
						comment = append(comment, data[offset:offset+end]...)
						offset = offset + end
					} else {
						return MalformedObjectError
					}
				}
			case '/':
				end := commentEnd(data[offset:])
				if end != -1 {
					if len(comment) > 0 {
						comment = append(comment, '\n')
					}
					comment = append(comment, data[offset:offset+end]...)
					offset = offset + end
				} else {
					return MalformedObjectError
				}
			default:
				endFlag = true
			}
			// 回调，这个时候注释解析好了
			flag, err := callback(key, value, valueType, offset, comment)
			if err != nil {
				return err
			}
			if !flag {
				return nil
			}
			comment = []byte{}

			endOff := nextToken(data[offset:])
			if endOff == -1 {
				return MalformedArrayError
			}
			offset += endOff

			if endFlag {
				switch data[offset] {
				case '}':
					return nil // Stop if we hit the close brace
				default:
					return MalformedObjectError
				}
			}
		}
	}

	return MalformedObjectError // we shouldn't get here; it's expected that we will return via finding the ending brace
}

// 数组迭代，如果flag为false，则停止迭代
func ArrayEach(data []byte, callback func(value []byte, dataType ValueType, offset int, comment []byte) (bool, error)) error {
	if len(data) == 0 {
		return MalformedObjectError
	}

	nT := nextToken(data)
	if nT == -1 {
		return MalformedJsonError
	}

	offset := nT + 1
	nO := nextToken(data[offset:])
	if nO == -1 {
		return MalformedJsonError
	}

	offset += nO

	if data[offset] == ']' {
		return nil
	}
	// 如果重复，只保留第一个
	comment := make([]byte, 0)
	for {
		nO = nextToken(data[offset:])
		if nO == -1 {
			return MalformedJsonError
		}
		offset += nO
		// 可能是注释，注释可能多行，循环解析，
		for data[offset] == '/' {
			end := commentEnd(data[offset:])
			if end != -1 {
				if len(comment) == 0 {
					comment = data[offset : offset+end]
				}
				offset = offset + end
				if off := nextToken(data[offset:]); off == -1 {
					return MalformedObjectError
				} else {
					offset += off
				}
			} else {
				return MalformedObjectError
			}
		}
		// 前面有多个注释，会在这里结束
		if data[offset] == ']' {
			break
		}

		v, t, o, e := Get(data[offset:])

		if e != nil {
			return e
		}

		if o == 0 {
			break
		}

		if t != NotExist {
			flag, err := callback(v, t, offset+o-len(v), comment)
			if err != nil {
				return err
			}
			if !flag {
				return nil
			}
		}

		if e != nil {
			break
		}

		offset += o

		skipToToken := nextToken(data[offset:])
		if skipToToken == -1 {
			return MalformedArrayError
		}
		offset += skipToToken

		// 可能是注释，注释可能多行，循环解析，
		for data[offset] == '/' {
			end := commentEnd(data[offset:])
			if end != -1 {
				if len(comment) == 0 {
					comment = data[offset : offset+end]
				}
				offset = offset + end
				if off := nextToken(data[offset:]); off == -1 {
					return MalformedObjectError
				} else {
					offset += off
				}
			} else {
				return MalformedObjectError
			}
		}

		if data[offset] == ']' {
			break
		}
		if data[offset] != ',' {
			return MalformedArrayError
		}
		offset++
	}
	return nil
}

// 判断是否是注释
func commentEnd(data []byte) int {
	for i := 1; i < len(data); i++ {
		if data[0] == '/' && data[1] == '/' && data[i] == '\n' {
			return i
		}
		if data[0] == '/' && data[1] == '*' && data[i-1] == '*' && data[i] == '/' {
			return i + 1
		}
	}
	return -1
}

func combineUTF16Surrogates(high, low rune) rune {
	return supplementalPlanesOffset + (high-highSurrogateOffset)<<10 + (low - lowSurrogateOffset)
}

func h2I(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	}
	return badHex
}

// decodeSingleUnicodeEscape decodes a single \uXXXX escape sequence. The prefix \u is assumed to be present and
// is not checked.
// In JSON, these escapes can either come alone or as part of "UTF16 surrogate pairs" that must be handled together.
// This function only handles one; decodeUnicodeEscape handles this more complex case.
func decodeSingleUnicodeEscape(in []byte) (rune, bool) {
	// We need at least 6 characters total
	if len(in) < 6 {
		return utf8.RuneError, false
	}

	// Convert hex to decimal
	h1, h2, h3, h4 := h2I(in[2]), h2I(in[3]), h2I(in[4]), h2I(in[5])
	if h1 == badHex || h2 == badHex || h3 == badHex || h4 == badHex {
		return utf8.RuneError, false
	}

	// Compose the hex digits
	return rune(h1<<12 + h2<<8 + h3<<4 + h4), true
}

// isUTF16EncodedRune checks if a rune is in the range for non-BMP characters,
// which is used to describe UTF16 chars.
// Source: https://en.wikipedia.org/wiki/Plane_(Unicode)#Basic_Multilingual_Plane
func isUTF16EncodedRune(r rune) bool {
	return highSurrogateOffset <= r && r <= basicMultilingualPlaneReservedOffset
}

func decodeUnicodeEscape(in []byte) (rune, int) {
	if r, ok := decodeSingleUnicodeEscape(in); !ok {
		// Invalid Unicode escape
		return utf8.RuneError, -1
	} else if r <= basicMultilingualPlaneOffset && !isUTF16EncodedRune(r) {
		// Valid Unicode escape in Basic Multilingual Plane
		return r, 6
	} else if r2, ok := decodeSingleUnicodeEscape(in[6:]); !ok { // Note: previous decodeSingleUnicodeEscape success guarantees at least 6 bytes remain
		// UTF16 "high surrogate" without manditory valid following Unicode escape for the "low surrogate"
		return utf8.RuneError, -1
	} else if r2 < lowSurrogateOffset {
		// Invalid UTF16 "low surrogate"
		return utf8.RuneError, -1
	} else {
		// Valid UTF16 surrogate pair
		return combineUTF16Surrogates(r, r2), 12
	}
}

// unescapeToUTF8 unescapes the single escape sequence starting at 'in' into 'out' and returns
// how many characters were consumed from 'in' and emitted into 'out'.
// If a valid escape sequence does not appear as a prefix of 'in', (-1, -1) to signal the error.
func unescapeToUTF8(in, out []byte) (inLen int, outLen int) {
	if len(in) < 2 || in[0] != '\\' {
		// Invalid escape due to insufficient characters for any escape or no initial backslash
		return -1, -1
	}

	// https://tools.ietf.org/html/rfc7159#section-7
	switch e := in[1]; e {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		// Valid basic 2-character escapes (use lookup table)
		out[0] = backslashCharEscapeTable[e]
		return 2, 1
	case 'u':
		// Unicode escape
		if r, inLen := decodeUnicodeEscape(in); inLen == -1 {
			// Invalid Unicode escape
			return -1, -1
		} else {
			// Valid Unicode escape; re-encode as UTF8
			outLen := utf8.EncodeRune(out, r)
			return inLen, outLen
		}
	}

	return -1, -1
}

// unescape unescapes the string contained in 'in' and returns it as a slice.
// If 'in' contains no escaped characters:
//
//	Returns 'in'.
//
// Else, if 'out' is of sufficient capacity (guaranteed if cap(out) >= len(in)):
//
//	'out' is used to build the unescaped string and is returned with no extra allocation
//
// Else:
//
//	A new slice is allocated and returned.
func Unescape(in, out []byte) ([]byte, error) {
	firstBackslash := bytes.IndexByte(in, '\\')
	if firstBackslash == -1 {
		return in, nil
	}

	// Get a buffer of sufficient size (allocate if needed)
	if cap(out) < len(in) {
		out = make([]byte, len(in))
	} else {
		out = out[0:len(in)]
	}

	// Copy the first sequence of unescaped bytes to the output and obtain a buffer pointer (subslice)
	copy(out, in[:firstBackslash])
	in = in[firstBackslash:]
	buf := out[firstBackslash:]

	for len(in) > 0 {
		// Unescape the next escaped character
		inLen, bufLen := unescapeToUTF8(in, buf)
		if inLen == -1 {
			return nil, MalformedStringEscapeError
		}

		in = in[inLen:]
		buf = buf[bufLen:]

		// Copy everything up until the next backslash
		nextBackslash := bytes.IndexByte(in, '\\')
		if nextBackslash == -1 {
			copy(buf, in)
			buf = buf[len(in):]
			break
		} else {
			copy(buf, in[:nextBackslash])
			buf = buf[nextBackslash:]
			in = in[nextBackslash:]
		}
	}

	// Trim the out buffer to the amount that was actually emitted
	return out[:len(out)-len(buf)], nil
}
