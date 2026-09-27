package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is the message we send to trigger the next step in the algorithm
type tickMsg time.Time

// tickEvery creates a command that fires a tickMsg every 800 milliseconds
func tickEvery() tea.Cmd {
	return tea.Tick(time.Millisecond*800, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model holds the state of our application
type model struct {
	s         string
	n         int
	pair      []int
	res       []byte
	i         int
	direction int
	done      bool
}

// initialModel sets up the wormhole pairs and the starting state
func initialModel() model {
	s := "(ed(et(oc))el)"
	n := len(s)
	pair := make([]int, n)
	var stack []int

	// Preprocess the pairs
	for i := 0; i < n; i++ {
		if s[i] == '(' {
			stack = append(stack, i)
		} else if s[i] == ')' {
			lastIdx := len(stack) - 1
			j := stack[lastIdx]
			stack = stack[:lastIdx]
			pair[i] = j
			pair[j] = i
		}
	}

	return model{
		s:         s,
		n:         n,
		pair:      pair,
		res:       []byte{},
		i:         0,
		direction: 1,
		done:      false,
	}
}

// Init runs when the TUI starts. We tell it to start ticking immediately.
func (m model) Init() tea.Cmd {
	return tickEvery()
}

// Update handles incoming events (like keyboard presses or time ticks)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keyboard inputs
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	// Handle our animation ticks
	case tickMsg:
		if m.done {
			return m, nil
		}

		// 1. Process the current character
		if m.s[m.i] == '(' || m.s[m.i] == ')' {
			// Teleport and reverse!
			m.i = m.pair[m.i]
			m.direction = -m.direction
		} else {
			// Add character to result
			m.res = append(m.res, m.s[m.i])
		}

		// 2. Move to the next index
		m.i += m.direction

		// 3. Check if we've exited the string bounds
		if m.i >= m.n || m.i < 0 {
			m.done = true
			return m, nil // Stop ticking
		}

		// Queue up the next tick
		return m, tickEvery()
	}

	return m, nil
}

// View defines how the UI is rendered to the terminal based on the current state
func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n  🐛 Wormhole Parentheses Visualizer 🐛\n\n")
	b.WriteString("  " + m.s + "\n")

	if !m.done {
		// Draw the moving pointer and direction indicator
		pointerSpace := strings.Repeat(" ", m.i)
		dirStr := "➡️  (Moving Right)"
		if m.direction == -1 {
			dirStr = "⬅️  (Moving Left)"
		}
		b.WriteString(fmt.Sprintf("  %s^ %s\n", pointerSpace, dirStr))
	} else {
		// Hide the pointer when finished
		b.WriteString("\n  ✨ Finished! ✨\n")
	}

	// Render the result array built so far
	b.WriteString(fmt.Sprintf("\n  Result: %s\n", string(m.res)))

	b.WriteString("\n  [Press 'q' to quit]\n\n")

	return b.String()
}

func main() {
	// Initialize and run the Bubble Tea program
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
