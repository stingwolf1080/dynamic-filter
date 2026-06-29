package types

const (

	// --- Validation ---
	ErrValidationMissingField = 1300 // Thiếu trường bắt buộc
	ErrValidationInvalidField = 1301 // Giá trị trường không hợp lệ
	ErrValidationDeviceID     = 1306 // Thiếu hoặc sai device ID

	// --- System / Internal ---
	ErrSystemDatabase      = 1400 // Lỗi truy vấn cơ sở dữ liệu
	ErrSystemTimeout       = 1401 // Quá thời gian xử lý
	ErrSystemParseFilter   = 1403 // Lỗi parse filter
	ErrSystemCache         = 1404 // Lỗi cache hoặc dữ liệu tạm
	ErrSystemConfigMissing = 1405 // Thiếu cấu hình hệ thống
	ErrOverlapData         = 1406 // overlap data
	ErrNotFoundData        = 1407 // data does not exists
	ErrTooManyRequest      = 1408 // to many request action
	ErrRequestCancelled    = 1409 // request cancelled
	ErrRequestProcessing   = 1410 // request is processing background job
)
