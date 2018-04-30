package xmpush

type Result struct {
	Result      string `json:"result"`
	TraceId     string `json:"trace_id"`
	Code        int64  `json:"code"`
	Description string `json:"description,omitempty"`
	Info        string `json:"info,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type SendResult struct {
	Result
	Data struct {
		ID string `json:"id,omitempty"`
	} `json:"data,omitempty"`
}

type Stat struct {
	Date                  string `json:"date"`
	AliasRecipients       int64  `json:"alias_recipients"`
	UserAccountRecipients int64  `json:"useraccount_recipients"`
	RegIDRecipients       int64  `json:"regid_recipients"`
	Received              int64  `json:"received"`
	BroadcastRecipients   int64  `json:"broadcast_recipients"`
	Click                 int64  `json:"click"`
	SingleRecipients      int64  `json:"single_recipients"`
}

type StatsResult struct {
	Result
	Data struct {
		Data []Stat `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

type Status struct {
	CreateTime      string `json:"create_time"`
	CreateTimestamp int64  `json:"create_timestamp"`
	TimeToLive      string `json:"time_to_live"`
	ClickRate       string `json:"click_rate"`
	MsgType         string `json:"msg_type"`
	DeliveryRate    string `json:"delivery_rate"`
	Delivered       int32  `json:"delivered"`
	ID              string `json:"id"`
	Click           int32  `json:"click"`
	Resolved        int32  `json:"resolved"`
}

type SingleStatusResult struct {
	Result
	Data struct {
		Data Status `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

type BatchStatusResult struct {
	Result
	Data struct {
		Data []Status `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

type InvalidRegIdsResult struct {
	Result
	Data struct {
		List []string `json:"list,omitempty"`
	} `json:"data,omitempty"`
}

type RegIdAalisResult struct {
	Result
	Data struct {
		List []string `json:"list,omitempty"`
	} `json:"data,omitempty"`
}

type RegIdTopicResult struct {
	Result
	Data struct {
		List []string `json:"list,omitempty"`
	} `json:"data,omitempty"`
}
