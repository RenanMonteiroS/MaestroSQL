package model

type ConnInfo struct {
	Host     string `json:"host" binding:"required"`
	Port     string `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	DbName   string `json:"dbName" binding:"required"`
}
