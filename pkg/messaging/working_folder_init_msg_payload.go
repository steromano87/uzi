package messaging

import "encoding/json"

const WorkingFolderInitMsgID = "WORKING_FOLDER_INIT"

type WorkingFolderInitPayload struct {
	CompressedWorkingFolder json.RawMessage `json:"compressedWorkingFolder"`
}

func NewWorkingFolderInitMsgID(zippedFolder []byte) Message {
	return NewRawMessage(WorkingFolderInitMsgID, &WorkingFolderInitPayload{CompressedWorkingFolder: zippedFolder})
}

func (w *WorkingFolderInitPayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string]*json.RawMessage

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}
	w.CompressedWorkingFolder = *intermediate[WorkingFolderInitMsgID]

	return nil
}
