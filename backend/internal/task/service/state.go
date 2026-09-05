package service

import "github.com/Yogdunana/StarByte/backend/internal/task/model"

// CanTransit reports whether status may move from -> to per Issue #9.
func CanTransit(from, to int16) bool {
	if from == to {
		return true
	}
	switch from {
	case model.StatusPending:
		return to == model.StatusDoing || to == model.StatusCancelled
	case model.StatusDoing:
		return to == model.StatusDone || to == model.StatusHeld || to == model.StatusCancelled
	case model.StatusHeld:
		return to == model.StatusDoing
	default:
		return false
	}
}

func canMutate(status int16) bool {
	return !model.IsClosed(status)
}
