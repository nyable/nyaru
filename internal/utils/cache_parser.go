package utils

import (
	"strings"

	"github.com/nyable/nyaru/internal/models"
)

// ParseScoopCacheOutput parses the raw text output of 'scoop cache show'
func ParseScoopCacheOutput(output string) []models.CacheResult {
	var results []models.CacheResult
	cleanOutput := strings.ReplaceAll(output, "\r\n", "\n")
	lines := strings.Split(cleanOutput, "\n")
	
	startParsing := false
	nameIdx, versionIdx, sizeIdx := -1, -1, -1
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Handle the case where scoop output is inside a JSON string array
		// e.g. "Name Version Size"
		// Trim brackets, quotes, and trailing commas
		line = strings.TrimLeft(line, " \"[,")
		line = strings.TrimRight(line, " \"],")
		if line == "" {
			continue
		}

		// Look for the header line to identify column indices
		if (strings.Contains(strings.ToLower(line), "name")) && 
		   (strings.Contains(strings.ToLower(line), "length") || 
		    strings.Contains(strings.ToLower(line), "size")) {
			headerFields := strings.Fields(strings.ToLower(line))
			for i, field := range headerFields {
				switch field {
				case "name":
					nameIdx = i
				case "version":
					versionIdx = i
				case "length", "size":
					sizeIdx = i
				}
			}
			if nameIdx != -1 && sizeIdx != -1 {
				startParsing = true
			}
			continue
		}

		// Look for the separator line
		if !startParsing && (strings.HasPrefix(line, "----") || strings.Contains(line, "----")) {
			if strings.Count(line, "-") > 5 {
				startParsing = true
				continue
			}
		}
		
		if !startParsing {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		
		// If this is the header again or separator, skip
		if strings.EqualFold(fields[0], "name") || strings.HasPrefix(fields[0], "---") {
			continue
		}

		// Use identified indices if possible, otherwise guess
		nIdx, vIdx, sIdx := nameIdx, versionIdx, sizeIdx
		if nIdx == -1 { nIdx = 0 }
		if vIdx == -1 { vIdx = 1 }
		// If we don't know the size index, it's often the 3rd field or the last one
		if sIdx == -1 {
			if len(fields) >= 3 {
				sIdx = 2
			} else {
				sIdx = len(fields) - 1
			}
		}

		if nIdx >= len(fields) || vIdx >= len(fields) || sIdx >= len(fields) {
			continue
		}

		name := fields[nIdx]
		version := fields[vIdx]
		
		// Size might be multiple fields (e.g. "1.90 MiB")
		sizeStr := fields[sIdx]
		// If the next field is a unit, include it
		if sIdx+1 < len(fields) {
			next := strings.ToUpper(fields[sIdx+1])
			isUnit := strings.Contains("B KB MB GB TB KIB MIB GIB TIB", next)
			if isUnit {
				sizeStr += " " + fields[sIdx+1]
			}
		}

		length := HumanSizeToBytes(sizeStr)
		if length == 0 && sIdx > 0 {
			// Try the field before if it's a number
			length = HumanSizeToBytes(fields[sIdx-1] + " " + fields[sIdx])
		}
		
		// Last effort: check any field that looks like a number + unit
		if length == 0 {
			for i := 0; i < len(fields)-1; i++ {
				l := HumanSizeToBytes(fields[i] + " " + fields[i+1])
				if l > 0 {
					length = l
					break
				}
			}
		}
		
		if (name != "" && !strings.Contains(name, "---")) && length > 0 {
			results = append(results, models.CacheResult{
				Name:       name,
				Version:    version,
				Length:     length,
				FormatSize: SizeToHumanRead(length),
			})
		}
	}
	
	return results
}

func stripAnsi(str string) string {
	// Simple manual ANSI stripper loop
	b := make([]byte, 0, len(str))
	inEscape := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (str[i] >= 'a' && str[i] <= 'z') || (str[i] >= 'A' && str[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		b = append(b, str[i])
	}
	return string(b)
}
