package recording

import (
	"fmt"
)

func (c *CapturedSession) GetMonitorName() string {
	return fmt.Sprintf("%s:%d:%t",c.Symbol, c.TimeFrame, c.IsFutures)
}
