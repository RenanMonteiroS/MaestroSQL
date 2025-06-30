package config

const (
	AuthenticatorUsage        = true                       // true/false
	AuthenticatorURL          = "http://192.168.1.66:8081" //The address to make calls to get a JWT
	AppHost                   = "0.0.0.0"                  //The host where the app will run. 0.0.0.0 to all addresses. If 0.0.0.0 is specified, the local IP is used for requests
	AppPort                   = 8881                       //The port where the app will run
	AppOpenOnceRunned         = true                       //An option to open the browser at the application address when the application is launched
	AppCertificateUsage       = false                      //If the HTTPs protocol will be used or not via certificate/key
	AppCertificateLocation    = ""                         //The location of the .crt/.pem file
	AppCertificateKeyLocation = ""                         //The location of the .key/.pem file
	AppCSRFTokenUsage         = true
	AppCSRFCookieSecret       = "my-supersecret-cookie"
	AppCSRFTokenSecret        = "my-supersecret-token"
)
