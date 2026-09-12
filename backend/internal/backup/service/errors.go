package service

import "github.com/Yogdunana/StarByte/backend/pkg/response"

func errNotFound() error {
	return response.NewError(response.CodeBackupNotFound, "备份记录不存在")
}

func errInvalidState(msg string) error {
	return response.NewError(response.CodeBackupInvalidState, msg)
}

func errBusy() error {
	return response.NewError(response.CodeBackupBusy, "已有备份或恢复任务在执行")
}

func errDump(msg string) error {
	return response.NewError(response.CodeBackupDumpFail, msg)
}

func errStore(msg string) error {
	return response.NewError(response.CodeBackupStoreFail, msg)
}

func errRestore(msg string) error {
	return response.NewError(response.CodeBackupRestoreFail, msg)
}

func errConfirm() error {
	return response.NewError(response.CodeBackupConfirmRequired, "恢复需 confirm=true 且 confirmation 输入 RESTORE")
}

func errDrillConfirm() error {
	return response.NewError(response.CodeBackupConfirmRequired, "演练需 confirm=true 且 confirmation 输入 DRILL")
}

func errPolicy(msg string) error {
	return response.NewError(response.CodeBackupPolicyInvalid, msg)
}

func errNotReady(msg string) error {
	return response.NewError(response.CodeBackupNotReady, msg)
}

func errChecksum() error {
	return response.NewError(response.CodeBackupChecksum, "备份校验和不匹配，已中止恢复")
}

func errDecrypt(msg string) error {
	return response.NewError(response.CodeBackupChecksum, msg)
}
