package service

import "github.com/Yogdunana/StarByte/backend/pkg/response"

func collectFail(msg string) error {
	return response.NewError(response.CodeMonitorCollectFail, msg)
}

func redisDown(msg string) error {
	return response.NewError(response.CodeMonitorRedisDown, msg)
}

func dbDown(msg string) error {
	return response.NewError(response.CodeMonitorDBDown, msg)
}
