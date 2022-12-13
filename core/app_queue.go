package core

import (
	"github.com/rokwire/logging-library-go/v2/logs"
)

type queueLogic struct {
	logger *logs.Logger

	storage Storage
}

func (q queueLogic) start() {
	q.logger.Info("queueLogic start")
}
