package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultString = "(ed(et(oc))el)"

// how many prior steps stay visible above the current one in the process box
const historyWindow = 4

type phase int

const (
	phaseInput phase = iota
	phaseViz
)

// segment is one leg of the traversal: either a run of reads that ends in a
// jump, or the final trailing run that ends when the walk exits the string.
type segment struct {
	track   string // arrow track for this leg, aligned under the string
	tunnel  string // wormhole span from fromIdx to toIdx, set only when isJump
	res     string // result built up through the end of this leg
	isJump  bool
	fromIdx int
	toIdx   int
	dir     int // direction (+1/-1) the walk was moving during this leg
}

// model holds the state of our application
type model struct {
	phase    phase
	inputBuf string
	inputErr string

	s        string
	segments []segment
	cursor   int
}

func initialModel() model {
	return model{phase: phaseInput}
}

func (m model) Init() tea.Cmd {
	return nil
}

func computePairs(s string) []int {
	pair := make([]int, len(s))
	var stack []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			stack = append(stack, i)
		case ')':
			last := len(stack) - 1
			j := stack[last]
			stack = stack[:last]
			pair[i] = j
			pair[j] = i
		}
	}
	return pair
}

func validateParens(s string) error {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return fmt.Errorf("unmatched ')' at position %d", i)
			}
		}
	}
	if depth != 0 {
		return fmt.Errorf("%d unclosed '('", depth)
	}
	return nil
}

// computeSegments replays the whole wormhole walk up front so the UI can
// step back and forth through it instead of only animating forward in time.
func computeSegments(s string, pair []int) []segment {
	n := len(s)
	track := make([]byte, n)
	for k := range track {
		track[k] = ' '
	}

	var segments []segment
	var res []byte
	i, direction := 0, 1

	for i >= 0 && i < n {
		isJump := s[i] == '(' || s[i] == ')'
		segDir := direction

		if isJump {
			track[i] = '^'
			to := pair[i]

			lo, hi := min(i, to), max(i, to)
			tunnel := make([]byte, n)
			for k := range tunnel {
				tunnel[k] = ' '
			}
			for k := lo + 1; k < hi; k++ {
				tunnel[k] = '~'
			}
			tunnel[lo] = '@'
			tunnel[hi] = '@'

			segments = append(segments, segment{
				track:   string(track),
				tunnel:  string(tunnel),
				res:     string(res),
				isJump:  true,
				fromIdx: i,
				toIdx:   to,
				dir:     segDir,
			})
			for k := range track {
				track[k] = ' '
			}
			i = to
			direction = -direction
		} else {
			if direction == 1 {
				track[i] = '>'
				if i > 0 && track[i-1] == '>' {
					track[i-1] = '-'
				}
			} else {
				track[i] = '<'
				if i < n-1 && track[i+1] == '<' {
					track[i+1] = '-'
				}
			}
			res = append(res, s[i])
		}

		i += direction
	}

	// Flush a trailing leg if the walk read characters after its last jump.
	if strings.TrimSpace(string(track)) != "" || len(segments) == 0 {
		segments = append(segments, segment{
			track:  string(track),
			res:    string(res),
			isJump: false,
			dir:    direction,
		})
	}

	return segments
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.phase {
	case phaseInput:
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			candidate := strings.TrimSpace(m.inputBuf)
			if candidate == "" {
				candidate = defaultString
			}
			if err := validateParens(candidate); err != nil {
				m.inputErr = err.Error()
				return m, nil
			}
			m.s = candidate
			m.segments = computeSegments(candidate, computePairs(candidate))
			m.cursor = 0
			m.inputErr = ""
			m.phase = phaseViz
		case tea.KeyBackspace:
			if r := []rune(m.inputBuf); len(r) > 0 {
				m.inputBuf = string(r[:len(r)-1])
			}
		case tea.KeySpace:
			m.inputBuf += " "
		case tea.KeyRunes:
			m.inputBuf += string(keyMsg.Runes)
		}

	case phaseViz:
		switch keyMsg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.segments)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g":
			m.cursor = 0
		case "G":
			m.cursor = len(m.segments) - 1
		case "r":
			m.phase = phaseInput
			m.inputBuf = ""
			m.inputErr = ""
		}
	}

	return m, nil
}

var (
	colorAccent   = lipgloss.Color("205")
	colorForward  = lipgloss.Color("51")
	colorBackward = lipgloss.Color("220")
	colorMuted    = lipgloss.Color("240")
	colorBorder   = lipgloss.Color("99")
	colorGreen    = lipgloss.Color("42")
	colorRed      = lipgloss.Color("196")

	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	boxStyle           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder).Padding(0, 1)
	boxTitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorBorder)
	hintStyle          = lipgloss.NewStyle().Foreground(colorMuted)
	errStyle           = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	mutedBlockStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	currentPrefixStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	jumpTagStyle       = lipgloss.NewStyle().Italic(true).Foreground(colorAccent)
	doneStyle          = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	forwardStyle       = lipgloss.NewStyle().Bold(true).Foreground(colorForward)
	backwardStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorBackward)
	jumpCharStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
)

// colorizeTrack colors '>'/'-'/'<' with dirStyle and '^' with jumpStyle,
// leaving spaces untouched. Only used for the current step so faded history
// blocks stay a single flat color.
func colorizeTrack(track string, dirStyle, jumpStyle lipgloss.Style) string {
	var b strings.Builder
	for _, ch := range track {
		switch ch {
		case '^':
			b.WriteString(jumpStyle.Render(string(ch)))
		case '>', '<', '-':
			b.WriteString(dirStyle.Render(string(ch)))
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// colorizeRunes wraps every rune found in targets with style, leaving the rest untouched.
func colorizeRunes(s string, style lipgloss.Style, targets string) string {
	var b strings.Builder
	for _, ch := range s {
		if strings.ContainsRune(targets, ch) {
			b.WriteString(style.Render(string(ch)))
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func renderSegment(s string, seg segment, current bool) string {
	if !current {
		lines := []string{"  " + s, "  " + seg.track}
		if seg.isJump {
			lines = append(lines, "  "+seg.tunnel, fmt.Sprintf("   jump %d → %d", seg.fromIdx, seg.toIdx))
		}
		return mutedBlockStyle.Render(strings.Join(lines, "\n"))
	}

	dirStyle := forwardStyle
	if seg.dir == -1 {
		dirStyle = backwardStyle
	}
	lines := []string{
		currentPrefixStyle.Render("▶ ") + s,
		"  " + colorizeTrack(seg.track, dirStyle, jumpCharStyle),
	}
	if seg.isJump {
		lines = append(lines,
			"  "+colorizeRunes(seg.tunnel, jumpCharStyle, "@~"),
			jumpTagStyle.Render(fmt.Sprintf("   jump %d → %d", seg.fromIdx, seg.toIdx)),
		)
	}
	return strings.Join(lines, "\n")
}

func (m model) renderResultBox(width int) string {
	last := len(m.segments) - 1
	progress := m.segments[m.cursor].res
	final := m.segments[last].res

	content := boxTitleStyle.Render("Result") + "\n" +
		fmt.Sprintf("so far: %q", progress)

	if m.cursor == last {
		content += "\n" + doneStyle.Render("✓ complete")
	} else {
		content += "\n" + hintStyle.Render(fmt.Sprintf("final:  %q  (step %d/%d)", final, m.cursor+1, len(m.segments)))
	}

	return boxStyle.Width(width).Render(content)
}

func (m model) renderProcessBox(width int) string {
	start := max(m.cursor-historyWindow+1, 0)

	var lines []string
	lines = append(lines, boxTitleStyle.Render("Process"))
	if start > 0 {
		lines = append(lines, hintStyle.Render(fmt.Sprintf("… %d earlier step(s) above", start)))
	}
	for idx := start; idx <= m.cursor; idx++ {
		if idx > start {
			lines = append(lines, "")
		}
		lines = append(lines, renderSegment(m.s, m.segments[idx], idx == m.cursor))
	}

	return boxStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m model) viewInput() string {
	display := m.inputBuf
	styledLine := display
	if display == "" {
		styledLine = hintStyle.Render(defaultString)
	}
	styledLine += "█"

	content := boxTitleStyle.Render("Input") + "\n" +
		"Enter a string with parentheses:\n\n" +
		styledLine
	if m.inputErr != "" {
		content += "\n\n" + errStyle.Render("✗ "+m.inputErr)
	}

	box := boxStyle.Width(50).Render(content)
	hint := hintStyle.Render("enter to confirm (blank = default) • esc/ctrl+c quit")

	return "\n" + titleStyle.Render("\U0001F573  Wormhole Parentheses Visualizer") + "\n\n" + box + "\n\n" + hint + "\n"
}

func (m model) viewViz() string {
	width := max(len(m.s)+8, 40)

	sections := []string{
		titleStyle.Render("\U0001F573  Wormhole Parentheses Visualizer"),
		"",
		m.renderResultBox(width),
		m.renderProcessBox(width),
		"",
		hintStyle.Render("j/k step • g/G first/last • r restart • q/ctrl+c quit"),
	}

	return "\n" + strings.Join(sections, "\n") + "\n"
}

func (m model) View() string {
	if m.phase == phaseInput {
		return m.viewInput()
	}
	return m.viewViz()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
