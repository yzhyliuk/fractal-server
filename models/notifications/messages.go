package notifications

import "fmt"

func ServerRestartedMessage() string {
	return fmt.Sprintf("Server was restarted. All active trades were closed. All strategy instances are stopped.")
}

func StrategyStopLoss(strategy string, totalLoss float64) string {
	return fmt.Sprintf("Your strategy %s was stopped due to stop loss condition. Total loss: is %f$",strategy,totalLoss)
}
