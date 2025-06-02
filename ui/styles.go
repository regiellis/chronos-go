package ui

import (
	"fmt"

	"github.com/regiellis/chronos-go/utils"
)

type ScaleIndicator struct {
	Active    bool
	Scale     string
	ScaleLeft int
}

func (s ScaleIndicator) View() string {
	if !s.Active {
		return ""
	}
	msg := fmt.Sprintf("[SCALE MODE] Next %d entries: %s", s.ScaleLeft, s.Scale)
	return utils.ActiveStyle.Render(msg)
}

// Example usage in a parent Bubble Tea view:
// scale := ui.ScaleIndicator{Active: true, Scale: "1h", ScaleLeft: 3}
// scale.View() // Renders the indicator
