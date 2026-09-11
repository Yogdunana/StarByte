package service

// 活动模块错误码统一使用 pkg/response 包中的常量（27000-27999）：
// CodeActivityNotFound = 27001
// CodeActivityInvalidState = 27002
// CodeActivityFull = 27003
// CodeRegistrationExists = 27004
// CodeRegistrationNotFound = 27005
// CodeCheckinFailed = 27006
// CodeCheckinAlreadyDone = 27007
// CodeCheckinNotApproved = 27008
// CodeSurveyAlreadySubmitted = 27009
// CodeSurveyNotEnded = 27010
// CodeCheckinTokenInvalid = 27011
// CodeCheckinGPSRejected = 27012
// CodeCheckinGPSNotConfigured = 27013
//
// 值班模块保留 26000-26999，互不占用。
