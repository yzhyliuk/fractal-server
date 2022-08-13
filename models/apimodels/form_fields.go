package apimodels

import "gorm.io/gorm"

const FieldsTableName = "form_fields"
const FormTableName = "forms"

type FormField struct {
	ID int `json:"id" gorm:"column:id"`
	FormID int `json:"form_id" gorm:"column:form_id"`
	Name string `json:"name" gorm:"column:name"`
	DisplayName string `json:"displayName" gorm:"column:display_name"`
	Description *string `json:"description" gorm:"column:description"`
	Min *int `json:"min" gorm:"column:min"`
	Max *int `json:"max" gorm:"column:max"`
	DefaultValue string `json:"defaultValue" gorm:"column:default_value"`
	Type string `json:"type" gorm:"column:type"`
	UiType string `json:"uiType" gorm:"column:ui_type"`
	Dataset string `json:"dataset" gorm:"column:dataset"`
}

type Form struct {
	ID int `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:id"`
}

func (f *Form) TableName() string {
	return FormTableName
}

func (f *FormField) TableName() string {
	return FieldsTableName
}

func GetFieldsByFormName(db *gorm.DB, formName string) ([]*FormField, error) {
	var form *Form
	err := db.Where("name = ?", formName).Find(&form).Error
	if err != nil {
		return nil, err
	}

	var fields []*FormField

	err = db.Where("form_id = ?", form.ID).Find(&fields).Error
	if err != nil {
		return nil, err
	}
	return fields, nil
}
