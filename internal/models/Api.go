package models

const (
	ApiPermView = 1 << iota
	ApiPermUpload
	ApiPermDelete
	ApiPermApiMod
	ApiPermEdit
)

const ApiPermNone = 0
const ApiPermAllNoApiMod = 23
const ApiPermAll = 31

// ApiKey contains data of a single api key
type ApiKey struct {
	Id             string `json:"Id"`
	FriendlyName   string `json:"FriendlyName"`
	LastUsedString string `json:"LastUsedString"`
	LastUsed       int64  `json:"LastUsed"`
	Permissions    uint8  `json:"Permissions"`
}

func (key *ApiKey) SetPermission(permission uint8) {
	key.Permissions |= permission
}
func (key *ApiKey) RemovePermission(permission uint8) {
	key.Permissions &^= permission
}

func (key *ApiKey) HasPermission(permission uint8) bool {
	if permission == ApiPermNone {
		return true
	}
	return (key.Permissions & permission) == permission
}

func (key *ApiKey) HasPermissionView() bool {
	return key.HasPermission(ApiPermView)
}

func (key *ApiKey) HasPermissionUpload() bool {
	return key.HasPermission(ApiPermUpload)
}

func (key *ApiKey) HasPermissionDelete() bool {
	return key.HasPermission(ApiPermDelete)
}

func (key *ApiKey) HasPermissionApiMod() bool {
	return key.HasPermission(ApiPermApiMod)
}

func (key *ApiKey) HasPermissionEdit() bool {
	return key.HasPermission(ApiPermEdit)
}
