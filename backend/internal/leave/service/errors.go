package service

import "github.com/Yogdunana/StarByte/backend/pkg/response"

func notFound() error {
	return response.NewError(response.CodeLeaveNotFound, "请假申请不存在")
}

func typeNotFound() error {
	return response.NewError(response.CodeLeaveTypeNotFound, "请假类型不存在")
}

func invalidState(msg string) error {
	return response.NewError(response.CodeLeaveInvalidState, msg)
}

func noAccess(msg string) error {
	return response.NewError(response.CodeLeaveNoAccess, msg)
}

func invalidTime(msg string) error {
	return response.NewError(response.CodeLeaveInvalidTime, msg)
}

func insufficient() error {
	return response.NewError(response.CodeLeaveInsufficient, "假期余额不足")
}

func overlap() error {
	return response.NewError(response.CodeLeaveOverlap, "该时间段已有请假申请")
}

func balanceMissing() error {
	return response.NewError(response.CodeLeaveBalanceMissing, "未找到该类型的假期余额")
}
