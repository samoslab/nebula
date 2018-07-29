package common

var (
	// NetworkUnreachable network cannot connected
	NetworkUnreachable = 99999
	// MaxInvalidDelay max time delay for provider
	MaxInvalidDelay = 1000

	// CC is concurrent
	CCDownloadGoNum = 5
	CCUploadGoNum   = 5
	CCUploadFileNum = 5

	TaskUploadFileType   = "UploadFile"
	TaskUploadDirType    = "UploadDir"
	TaskDownloadFileType = "DownloadFile"
	TaskDownloadDirType  = "DownloadDir"

	TaskUploadProgressType   = "UploadProgress"
	TaskDownloadProgressType = "DownloadProgress"

	MsgQueueLen  = 1000
	TaskQuqueLen = 1000
)
