// Package store 提供业务数据访问层。
package store

import (
	"aitrendings/pkg/db"
)

// Store 业务数据存储
type Store struct {
	db *db.Database
}
