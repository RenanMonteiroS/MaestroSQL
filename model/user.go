package model

type OAuth2User struct {
	GoogleOAuth2User    GoogleOAuth2User
	MicrosoftOAuth2User MicrosoftOAuth2User
}

type MicrosoftOAuth2User struct {
	DataContext       string
	BusinessPhones    []string
	DisplayName       string
	GivenName         string
	JobTitle          *string
	Mail              *string
	MobilePhone       *string
	OfficeLocation    *string
	PreferredLanguage string
	Surname           string
	UserPrincipalName string
	Id                string
}

type GoogleOAuth2User struct {
	Id         string
	Email      string
	FamilyName string
	Picture    string
}
