package xiaolongbaopool

import "errors"

var (
	ErrInvalidPoolSize   = errors.New("invalid pool size")
	ErrInvalidPoolExpiry = errors.New("invalid pool expiry time")
	ErrPoolClosed        = errors.New("pool closed")
)

const (
	DEFAULT_CLEAN_DURATION = 1
)
