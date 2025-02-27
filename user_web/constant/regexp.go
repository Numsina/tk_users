package constant

const UserEmail = "\\w[-\\w.+]*@([A-Za-z0-9][-A-Za-z0-9]+\\.)+[A-Za-z]{2,14}"
const UserPassword = "^(?![a-zA-Z]+$)(?!\\d+$)(?![^\\da-zA-Z\\s]+$).{6,16}$"
const NickName = "^[a-zA-Z]\\w{4,15}$"
const BirthDay = "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$"
const AbortMe = "^(?:[a-zA-Z]{0,255})$"
const PhoneNumber = "/^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\\d{8}$/"
