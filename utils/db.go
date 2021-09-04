package utils

import (
	"net/http"

	"gorm.io/gorm"
)

func DefaultFindFromRequest(r *http.Request, db *gorm.DB, data interface{}) error {
	id, err := ReadIntURL(r, "id")
	if err != nil {
		return err
	}
	if err := db.Set("gorm:auto_preload", true).First(data, id).Error; err != nil {
		return err
	}
	return nil
}

func DefaultQueryAll(r *http.Request, q *gorm.DB) *gorm.DB {
	return q.Set("gorm:auto_preload", true)
}
