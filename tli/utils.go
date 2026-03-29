package tli

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func centerText(width int, text string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		padding := (width - visibleWidth(line)) / 2
		if padding > 0 {
			lines[i] = strings.Repeat(" ", padding) + line
		}
	}

	return strings.Join(lines, "\n")
}

func centerBlock(width int, text string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if visibleWidth(line) > maxLen {
			maxLen = visibleWidth(line)
		}
	}

	padding := (width - maxLen) / 2
	if padding < 0 {
		padding = 0
	}

	prefix := strings.Repeat(" ", padding)
	for i, line := range lines {
		lines[i] = prefix + line
	}

	return strings.Join(lines, "\n")
}

func ansi(code string, text string) string {
	return "\033[" + code + "m" + text + "\033[0m"
}

func muted(text string) string {
	return ansi("38;5;245", text)
}

func accent(text string) string {
	return ansi("38;5;81", text)
}

func highlight(text string) string {
	return ansi("38;5;229", text)
}

func success(text string) string {
	return ansi("38;5;120", text)
}

func danger(text string) string {
	return ansi("38;5;210", text)
}

func strong(text string) string {
	return ansi("1", text)
}

func box(title string, lines []string) string {
	width := visibleWidth(title) + 4
	for _, line := range lines {
		if visibleWidth(line) > width-4 {
			width = visibleWidth(line) + 4
		}
	}

	top := "┌" + strings.Repeat("─", width-2) + "┐"
	header := "│ " + strong(title) + strings.Repeat(" ", width-visibleWidth(title)-3) + "│"
	body := make([]string, 0, len(lines)+3)
	body = append(body, top, header, "├"+strings.Repeat("─", width-2)+"┤")
	for _, line := range lines {
		padding := width - visibleWidth(line) - 3
		if padding < 0 {
			padding = 0
		}
		body = append(body, "│ "+line+strings.Repeat(" ", padding)+"│")
	}
	body = append(body, "└"+strings.Repeat("─", width-2)+"┘")
	return strings.Join(body, "\n")
}

func joinColumns(left string, right string, gap int) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	maxLines := max(len(leftLines), len(rightLines))
	leftWidth := 0
	for _, line := range leftLines {
		if visibleWidth(line) > leftWidth {
			leftWidth = visibleWidth(line)
		}
	}

	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	joined := make([]string, 0, maxLines)
	spacer := strings.Repeat(" ", gap)
	for i := 0; i < maxLines; i++ {
		joined = append(joined, padVisibleRight(leftLines[i], leftWidth)+spacer+rightLines[i])
	}

	return strings.Join(joined, "\n")
}

func padVisibleRight(text string, width int) string {
	padding := width - visibleWidth(text)
	if padding <= 0 {
		return text
	}
	return text + strings.Repeat(" ", padding)
}

func blockWidth(text string) int {
	lines := strings.Split(text, "\n")
	width := 0
	for _, line := range lines {
		if visibleWidth(line) > width {
			width = visibleWidth(line)
		}
	}
	return width
}

func visibleWidth(text string) int {
	width := 0
	for i := 0; i < len(text); {
		if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) && text[i] != 'm' {
				i++
			}
			if i < len(text) {
				i++
			}
			continue
		}

		_, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 {
			break
		}
		width++
		i += size
	}

	return width
}

func parseFloat(value string, label string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", label)
	}

	out, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}

	if out < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}

	return out, nil
}

func parseInt(value string, label string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", label)
	}

	out, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number", label)
	}

	if out < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}

	return out, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "y", "yes", "true":
		return true, nil
	case "n", "no", "false":
		return false, nil
	default:
		return false, fmt.Errorf("off fridays must be yes or no")
	}
}
