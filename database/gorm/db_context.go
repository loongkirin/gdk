package gorm

import (
	"gorm.io/gorm"
)

type DbContext interface {
	GetMasterDb() *gorm.DB
	GetSlaveDb() *gorm.DB
}
