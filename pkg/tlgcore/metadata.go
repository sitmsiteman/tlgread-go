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

			// Decode the ID string to get the work number (e.g. "001")
			idStr := decodeWorkID(idBytes)
			currentWork = &WorkMetadata{ID: idStr}
			m[idStr] = currentWork

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
			// Unknown Type code.
			// fmt.Printf("Debug: Unknown IDT Opcode %02x at offset %d\n", typ, pos-1)
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

func decodeWorkID(b []byte) string {
	idx := 0
	if len(b) >= 2 && b[0] == 0xEF && b[1] == 0x81 {
		idx = 2
	}

	var sb strings.Builder
	for i := idx; i < len(b); i++ {
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
