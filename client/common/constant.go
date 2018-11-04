package common

var (
	// NetworkUnreachable network cannot connected
	NetworkUnreachable = 99999
	// MaxInvalidDelay max time delay for provider
	MaxInvalidDelay = 100

	// CC is concurrent
	CCDownloadGoNum = 3
	CCUploadGoNum   = 3
	CCUploadFileNum = 3
	CCTaskHandleNum = 3

	TaskUploadFileType   = "UploadFile"
	TaskUploadDirType    = "UploadDir"
	TaskDownloadFileType = "DownloadFile"
	TaskDownloadDirType  = "DownloadDir"

	TaskUploadProgressType   = "UploadProgress"
	TaskDownloadProgressType = "DownloadProgress"

	MsgQueueLen  = 1000
	TaskQuqueLen = 1000
	MetaQuqueLen = 1000
)
