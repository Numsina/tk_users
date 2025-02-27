package domain

type User struct {
	Id              int32  `json:"id"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	NickName        string `json:"nick_name"`
	Description     string `json:"description"`
	Avatar          string `json:"avatar"`
	BirthDay        int64  `json:"birth_day"`
	Address         string `json:"address"`
}

type UserResp struct {
	Id          int32  `json:"id"`
	Email       string `json:"email"`
	NickName    string `json:"nick_name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	BirthDay    int64  `json:"birth_day"`
	Address     string `json:"address"`
	CreateAt    int64  `json:"create_at"`
}
