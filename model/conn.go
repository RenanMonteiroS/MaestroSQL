package model

import "log/slog"

// ConnInfo is a set of Host, Port, User, Password and MaxConnPool. Its populated by JSON, via HTTP request. Expects a host, port, user and password in the request body.
type ConnInfo struct {
	Host                   string `json:"host" binding:"required"`
	Port                   string `json:"port"`
	User                   string `json:"user" binding:"required"`
	Password               string `json:"password" binding:"required"`
	Instance               string `json:"instance"`
	Encryption             string `json:"encryption"`
	MaxConnections         int    `json:"maxConnections"`
	TrustServerCertificate *bool  `json:"trustServerCertificate"`
}

func (ci ConnInfo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", ci.Host),
		slog.String("port", ci.Port),
		slog.String("user", ci.User),
		slog.String("password", "REDACTED"),
		slog.String("instance", ci.Instance),
		slog.Int("maxConnections", ci.MaxConnections),
	)
}
