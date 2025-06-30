package config

const (
	AuthenticatorUsage        = true                       // true/false
	AuthenticatorURL          = "http://192.168.1.66:8081" //The address to make calls to get a JWT
	AppHost                   = "0.0.0.0"                  //The host where the app will run. 0.0.0.0 to all addresses
	AppPort                   = 8881                       //The port where the app will run
	AppOpenOnceRunned         = true                       //An option to open the browser at the application address when the application is launched
	AppCertificateUsage       = false
	AppCertificateLocation    = ""
	AppCertificateKeyLocation = ""
)
