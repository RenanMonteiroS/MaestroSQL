package model

// ConnInfo is a set of Host, Port, User and Password. Its populated by JSON, via HTTP request. Expects a host, port, user and password in the request body.
type ConnInfo struct {
	Host        string `json:"host" binding:"required"`
	Port        string `json:"port" binding:"required"`
	User        string `json:"user" binding:"required"`
	Password    string `json:"password" binding:"required"`
	MaxConnPool string `json:"maxConnPool"`
}
