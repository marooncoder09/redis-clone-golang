package utils

const (
	MasterReplID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb" // TODO: even though this is hardcoded rn, we should pick this from the .env file instead
)

func GetMasterReplID() string {
	return MasterReplID
}
