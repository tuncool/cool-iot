package models

func GetAllTable() []interface{} {
	return []interface{}{AuthUser{}, Group{}, Client{}, Config{}, Log{}}
}
