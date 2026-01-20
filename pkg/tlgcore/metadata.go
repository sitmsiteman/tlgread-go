package tlgcore

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type CitationDef struct {
	LevelChar string // "v", "w", "x", "y", "z"
	Label     string // e.g. "Book", "Line"
}

type WorkMetadata struct {
	ID        string
	Title     string
	Citations []CitationDef
}

func ReadIDT(path string) (map[string]*WorkMetadata, error) {
	m := make(map[string]*WorkMetadata)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pos := 0
	var currentWork *WorkMetadata

	// consumeID reads a sequence of bytes as long as they have the high bit set.
	// TLG IDT files use high-bit bytes for ID data to distinguish from Type codes.
	consumeID := func() []byte {
		start := pos
		for pos < len(data) && data[pos] >= 0x80 {
			pos++
		}
		return data[start:pos]
	}

	for pos < len(data) {
		// Read Type Code
		typ := data[pos]
		pos++

		switch typ {
		case 0: // End of File (or Padding)
			continue

		case 1: // New Author (Type 1)
			// Format: 01 [Len:2] [Block:2] [ID String...]
			if pos+4 > len(data) {
				break
			}
			pos += 4 // Skip Len and Block
			consumeID()

		case 2: // New Work (Type 2)
			// Format: 02 [Len:2] [Block:2] [ID String...]
			if pos+4 > len(data) {
				break
			}
			pos += 4 // Skip Len and Block

			// The ID String here contains the Work ID (Level b)
			idBytes := consumeID()

			idStr := DecodeWorkID(idBytes)
			currentWork = &WorkMetadata{ID: idStr}

			if idStr != "" {
				m[idStr] = currentWork
			}

		case 3: // New Section (Type 3)
			// Format: 03 [Block:2]
			if pos+2 > len(data) {
				break
			}
			pos += 2

		case 8, 9, 10, 12, 13: // ID Fields
			consumeID()

		case 11: // Start Exception (Type 11)
			// Format: 11 [Block:2] [ID String...]
			if pos+2 > len(data) {
				break
			}
			pos += 2
			consumeID()

		case 16: // Description of ID fields a,b (Type 16)
			// Format: 10 [Subtype:1] [Len:1] [String...]
			if pos+2 > len(data) {
				break
			}
			subtype := data[pos]
			pos++
			length := int(data[pos])
			pos++

			if pos+length > len(data) {
				break
			}
			str := string(data[pos : pos+length])
			pos += length

			// Subtype 1 = Work Title
			if subtype == 1 && currentWork != nil {
				currentWork.Title = cleanString(str)
			}

		case 17: // Description of ID fields v..z (Type 17)
			// Format: 11 [Subtype:1] [Len:1] [String...]
			if pos+2 > len(data) {
				break
			}
			subtype := data[pos]
			pos++
			length := int(data[pos])
			pos++

			if pos+length > len(data) {
				break
			}
			str := string(data[pos : pos+length])
			pos += length

			if currentWork != nil {
				levelChar := ""
				switch subtype {
				case 4:
					levelChar = "v"
				case 3:
					levelChar = "w"
				case 2:
					levelChar = "x"
				case 1:
					levelChar = "y"
				case 0:
					levelChar = "z"
				}
				if levelChar != "" {
					currentWork.Citations = append(currentWork.Citations, CitationDef{
						LevelChar: levelChar,
						Label:     cleanString(str),
					})
				}
			}

		default:
			// Unknown Type code or sync error
			continue
		}
	}
	return m, nil
}

func cleanString(s string) string {
	if strings.Contains(s, "*") {
		return ToGreek(s)
	}
	return ToLatin(s)
}

// DecodeWorkID parses the binary ID bytes to extract the Work ID (Level b).
// It handles standard binary tags (Escape 0xE -> Level b) and extracts values.
func DecodeWorkID(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	// 1. Check for legacy ASCII string format (prefixed with EF 81)
	if len(b) >= 2 && b[0] == 0xEF && b[1] == 0x81 {
		return decodeSimpleASCII(b[2:])
	}

	// 2. Binary Parser state
	pos := 0

	readBin := func(n int) int {
		v := 0
		for i := 0; i < n; i++ {
			if pos >= len(b) {
				break
			}
			v = (v << 7) | int(b[pos]&0x7F)
			pos++
		}
		return v
	}

	readStr := func() string {
		var sb strings.Builder
		for pos < len(b) {
			val := b[pos]
			// String ends at 0xFF or end of slice
			if val == 0xFF {
				pos++
				break
			}
			sb.WriteByte(val & 0x7F)
			pos++
		}
		return sb.String()
	}

	// Scan through the bytes looking for Level "b" assignment
	for pos < len(b) {
		val := b[pos]
		pos++

		left := (val >> 4) & 0x0F
		right := val & 0x0F

		isLevelB := false

		// Determine Level
		// Level 'b' is indicated by Escape (0xE) followed by 0x81 (which is 1 | 0x80)
		if left == 0xE {
			if pos < len(b) {
				next := b[pos] & 0x7F
				pos++
				if next == 1 {
					isLevelB = true
				}
			}
		}

		// Decode Value
		var numVal int = -999
		var strVal string

		switch right {
		case 0x0:
			numVal = -1 // Auto-increment signal (unlikely for explicit IDT)
		case 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7:
			numVal = int(right)
		case 0x8:
			numVal = readBin(1)
		case 0x9:
			numVal = readBin(1)
			strVal = string(readBin(1)) // effectively readChar
		case 0xA:
			numVal = readBin(1)
			strVal = readStr()
		case 0xB:
			numVal = readBin(2)
		case 0xC:
			numVal = readBin(2)
			strVal = string(readBin(1))
		case 0xD:
			numVal = readBin(2)
			strVal = readStr()
		case 0xE:
			strVal = string(readBin(1)) // readChar
		case 0xF:
			strVal = readStr()
		}

		// If this was a Level B tag, return the value immediately
		if isLevelB {
			if strVal != "" {
				if numVal != -999 && numVal != -1 {
					return strconv.Itoa(numVal) + strVal
				}
				return strVal
			}
			if numVal != -999 {
				return strconv.Itoa(numVal)
			}
		}
	}

	// 3. Fallback: If no binary tag for 'b' was found, try naive ASCII.
	// This handles cases where ID is just a string without EF 81 prefix.
	return decodeSimpleASCII(b)
}

func decodeSimpleASCII(b []byte) string {
	var sb strings.Builder
	for i := 0; i < len(b); i++ {
		if b[i] == 0xFF {
			break
		}
		if b[i] >= 0x80 {
			val := b[i] & 0x7F
			if (val >= '0' && val <= '9') || (val >= 'A' && val <= 'Z') || (val >= 'a' && val <= 'z') {
				sb.WriteByte(val)
			}
		}
	}
	res := sb.String()
	if i, err := strconv.Atoi(res); err == nil {
		return strconv.Itoa(i)
	}
	return res
}

func GetAuthorName(path, tlgID string) string {
	var prefixID string
	data, err := os.ReadFile(path)
	if err != nil {
		return "Unknown"
	}

	if len(tlgID) >= 3 {
		prefixID = strings.ToUpper(tlgID[:3])
	} else {
		return "Unknown"
	}

	cleanID := fmt.Sprintf("%s%04s", prefixID, strings.TrimPrefix(strings.ToUpper(tlgID), prefixID))
	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?&1(.*?)&`, cleanID))
	matches := re.FindSubmatch(data)
	if len(matches) > 1 {
		return strings.TrimSpace(strings.Split(string(matches[1]), "&")[0])
	}
	return tlgID
}
