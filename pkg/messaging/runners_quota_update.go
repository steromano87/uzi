package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
)

const RunnersQuotaUpdateMsgId = "RUNNERS_QUOTA_UPDATE"

type RunnersQuotaUpdatePayload struct {
	Quota int `json:"quota"`
}

func NewRunnerQuotaUpdateMessage(runnerQuota int) Message {
	return NewRawMessage(RunnersQuotaUpdateMsgId, &RunnersQuotaUpdatePayload{Quota: runnerQuota})
}

func (r *RunnersQuotaUpdatePayload) Type() string {
	return RunnersQuotaUpdateMsgId
}

func (r *RunnersQuotaUpdatePayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string]any

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}

	quota, ok := intermediate[RunnersQuotaUpdateMsgId].(int)
	if !ok {
		return errors.New(fmt.Sprintf("error parsing runner '%s' field as integer: %x", RunnersQuotaUpdateMsgId, quota))
	}

	r.Quota = quota
	return nil
}
