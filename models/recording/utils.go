package recording

import (
	"fmt"
)

func (c *CapturedSession) GetMonitorName() string {
	return fmt.Sprintf("%s:%d", c.Symbol, c.TimeFrame)
}
