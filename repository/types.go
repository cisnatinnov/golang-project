package repository

type Estate struct {
	Id     string
	Length int
	Width  int
}

type Tree struct {
	Id       string
	EstateId string
	X        int
	Y        int
	Height   int
}

type CreateEstateInput struct {
	Length int
	Width  int
}

type CreateEstateOutput struct {
	Id string
}

type CreateTreeInput struct {
	EstateId string
	X        int
	Y        int
	Height   int
}

type CreateTreeOutput struct {
	Id string
}

type GetEstateStatsInput struct {
	EstateId string
}

type GetEstateStatsOutput struct {
	Count  int
	Max    int
	Min    int
	Median int
}

type GetTreesByEstateIdInput struct {
	EstateId string
}

type GetTreesByEstateIdOutput struct {
	Trees []Tree
}

type User struct {
	Id           string
	Username     string
	PasswordHash string
}

type CreateUserInput struct {
	Username     string
	PasswordHash string
}

type CreateUserOutput struct {
	Id string
}

type UpdateUserInput struct {
	Id           string
	Username     string
	PasswordHash string
}

type GetUserByUsernameOrEmailInput struct {
	Username string
	Email    string
}

type Person struct {
	Id          string
	UserId      string
	FirstName   *string
	LastName    *string
	DateOfBirth *string // DATE format: YYYY-MM-DD
	Bio         *string
	AvatarUrl   *string
}

type CreatePersonInput struct {
	UserId      string
	FirstName   *string
	LastName    *string
	DateOfBirth *string
	Bio         *string
	AvatarUrl   *string
}

type CreatePersonOutput struct {
	Id string
}

type UpdatePersonInput struct {
	Id          string
	FirstName   *string
	LastName    *string
	DateOfBirth *string
	Bio         *string
	AvatarUrl   *string
}

type GetPersonByUserIdInput struct {
	UserId string
}

// Email entities
type PersonEmail struct {
	Id        string
	UserId    string
	Email     string
	IsPrimary bool
	Verified  bool
}

type CreatePersonEmailInput struct {
	UserId    string
	Email     string
	IsPrimary bool
}

type CreatePersonEmailOutput struct {
	Id string
}

type UpdatePersonEmailInput struct {
	Id        string
	IsPrimary *bool
	Verified  *bool
}

type GetPersonEmailsByUserIdInput struct {
	UserId string
}

type GetPersonEmailsByUserIdOutput struct {
	Emails []PersonEmail
}

// Phone entities
type PersonPhone struct {
	Id        string
	UserId    string
	Phone     string
	Type      *string // 'mobile', 'home', 'work'
	IsPrimary bool
	Verified  bool
}

type CreatePersonPhoneInput struct {
	UserId    string
	Phone     string
	Type      *string
	IsPrimary bool
}

type CreatePersonPhoneOutput struct {
	Id string
}

type UpdatePersonPhoneInput struct {
	Id        string
	Type      *string
	IsPrimary *bool
	Verified  *bool
}

type GetPersonPhonesByUserIdInput struct {
	UserId string
}

type GetPersonPhonesByUserIdOutput struct {
	Phones []PersonPhone
}

// Social Media entities
type PersonSocialMedia struct {
	Id         string
	UserId     string
	Platform   string // 'twitter', 'linkedin', 'github', etc.
	Username   *string
	ProfileUrl *string
}

type CreatePersonSocialMediaInput struct {
	UserId     string
	Platform   string
	Username   *string
	ProfileUrl *string
}

type CreatePersonSocialMediaOutput struct {
	Id string
}

type UpdatePersonSocialMediaInput struct {
	Id         string
	Username   *string
	ProfileUrl *string
}

type GetPersonSocialMediaByUserIdInput struct {
	UserId string
}

type GetPersonSocialMediaByUserIdOutput struct {
	SocialMediaAccounts []PersonSocialMedia
}
