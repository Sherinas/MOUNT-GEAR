package models

import (
	"strconv"

	"gorm.io/gorm"
)

func FetchData(db *gorm.DB, model interface{}) error {
	if err := db.Find(model).Error; err != nil {
		return err
	}
	return nil
}

func CheckExists(db *gorm.DB, model interface{}, condition string, args ...interface{}) bool {
	return db.Where(condition, args...).First(model).Error == nil
}

func EmailExists(db *gorm.DB, email string, user *User) error {
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}
	return nil
}

func CreateRecord(db *gorm.DB, model interface{}, data interface{}) error {
	if err := db.Create(data).Error; err != nil {
		return err
	}
	return nil
}

func GetRecordByID(db *gorm.DB, model interface{}, id string) error {
	// Convert id to integer
	recordID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	// Fetch the record by ID
	if err := db.First(model, recordID).Error; err != nil {
		return err
	}

	return nil
}

func UpdateRecord(db *gorm.DB, model interface{}) error {
	if err := db.Save(model).Error; err != nil {
		return err
	}
	return nil
}

func SearchRecord(db *gorm.DB, query string, datas *[]Category) error {
	if err := db.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").Find(datas).Error; err != nil {
		return err
	}
	return nil
}

func CheckStatus(db *gorm.DB, condition interface{}, model interface{}) error {
	if err := db.Where("is_active = ?", condition).Find(model).Error; err != nil {
		return err
	}
	return nil
}

func UpdateModel(db *gorm.DB, model interface{}, updates map[string]interface{}) error {
	if err := db.Model(model).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func FindUserByID(db *gorm.DB, id string, user *User) error {
	if err := db.First(user, id).Error; err != nil {
		return err
	}
	return nil
}
