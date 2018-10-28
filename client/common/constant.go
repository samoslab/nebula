package common

var (
	// NetworkUnreachable network cannot connected
	NetworkUnreachable = 99999
	// MaxInvalidDelay max time delay for provider
	MaxInvalidDelay = 100

	// CC is concurrent
	CCDownloadGoNum = 1
	CCUploadGoNum   = 1
	CCUploadFileNum = 1
	CCTaskHandleNum = 1

	TaskUploadFileType   = "UploadFile"
	TaskUploadDirType    = "UploadDir"
	TaskDownloadFileType = "DownloadFile"
	TaskDownloadDirType  = "DownloadDir"

	TaskUploadProgressType   = "UploadProgress"
	TaskDownloadProgressType = "DownloadProgress"

	MsgQueueLen  = 1000
	TaskQuqueLen = 1000
)
