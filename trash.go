package tree

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type TrashInfo struct {
	OriginalPath string
	CreatedAt    time.Time
}

func NewTrashInfo(originalPath string) TrashInfo {
	return TrashInfo{
		OriginalPath: originalPath,
		CreatedAt:    time.Now(),
	}
}

func DecodeTrashInfo(str string) (TrashInfo, error) {
	var t TrashInfo
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(b, &t); err != nil {
		return t, err
	}
	return t, nil
}

func (t TrashInfo) Encode() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
