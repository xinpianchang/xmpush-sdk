package xmpush

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"testing"
	"time"
)

var (
	/*
	测试参数, 需要自行建立配置文件 test_data.json
	配置范例

		{
		  "appSecret": "appSecret",
		  "packageName": [
			"com.server.example"
		  ],
		  "regId": [
			"regid_1"
		  ],
		  "alias": [
			"alias_1"
		  ],
		  "account": [
			"account_1"
		  ],
		  "topic": [
			"topic_1"
		  ]
		}

	*/
	appSecret   string
	packageName []string
	regId       []string
	alias       []string
	account     []string
	topic       []string

	client *Client
	l      = newSimpleLogger()

	message = NewMessage("hello world", "hello world description")
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	dataFile := path.Join(path.Dir(file), "test_data.json")
	bytes, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Fatal("read test_data error", err)
	}

	var data struct {
		AppSecret   string   `json:"appSecret"`
		PackageName []string `json:"packageName"`
		RegId       []string `json:"regId"`
		Alias       []string `json:"alias"`
		Account     []string `json:"account"`
		Topic       []string `json:"topic"`
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		log.Fatal("parse test_data json error", err)
	}

	if data.AppSecret == "" ||
		data.PackageName == nil ||
		data.RegId == nil ||
		data.Alias == nil ||
		data.Account == nil {
		log.Fatal("test_data invalid")
	}

	appSecret = data.AppSecret
	packageName = data.PackageName
	regId = data.RegId
	alias = data.Alias
	account = data.Account
	topic = data.Topic

	client, err = NewClient(appSecret, packageName...)
	if err != nil {
		log.Fatal(err)
	}
	client.SetLogger(l)

	message.SetBadge(10)

	// 是否启用沙箱测试，沙箱偶尔不稳定，返回错误
	//client.UseSandbox(true)
}

func TestClient_SendToRegId(t *testing.T) {
	result, err := client.SendToRegId(message, &regId)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_SendToAlias(t *testing.T) {
	result, err := client.SendToAlias(message, &alias)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_SendToAccount(t *testing.T) {
	result, err := client.SendToAccount(message, &account)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_SendToTopic(t *testing.T) {
	result, err := client.SendToTopic(message, topic[0])
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_SendToTopics(t *testing.T) {
	result, err := client.SendToTopics(message, &topic, "")
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_SendTargetedMessage(t *testing.T) {

	// 沙箱不成功
	client.UseSandbox(false)

	{
		targetedMessages := []TargetedMessage{
			*NewTargetedMessage(NewMessage("target regId", "description"), regId[0], TargetRegId),
		}
		result, err := client.SendTargetedMessage(&targetedMessages)
		if err != nil {
			t.Fatal(err)
		}

		if result.Code != 0 {
			t.Fatal(result.Code, result.Description)
		}

		l.Debug(result.Data.ID)
		l.Debug(result.Info)
	}

	{
		targetedMessages := []TargetedMessage{
			*NewTargetedMessage(NewMessage("target alias", "description"), alias[0], TargetAlias),
		}
		result, err := client.SendTargetedMessage(&targetedMessages)
		if err != nil {
			t.Fatal(err)
		}

		if result.Code != 0 {
			t.Fatal(result.Code, result.Description)
		}

		l.Debug(result.Data.ID)
		l.Debug(result.Info)
	}

	{
		targetedMessages := []TargetedMessage{
			*NewTargetedMessage(NewMessage("target account", "description"), account[0], TargetAccount),
		}
		result, err := client.SendTargetedMessage(&targetedMessages)
		if err != nil {
			t.Fatal(err)
		}

		if result.Code != 0 {
			t.Fatal(result.Code, result.Description)
		}

		l.Debug(result.Data.ID)
		l.Debug(result.Info)
	}

}

func TestClient_SendToAll(t *testing.T) {
	result, err := client.SendToAll(message)
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.ID)
	l.Debug(result.Info)
}

func TestClient_Stats(t *testing.T) {
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()
	format := "20060102"
	result, err := client.Stats(start.Format(format), end.Format(format))

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	stats, _ := json.MarshalIndent(&result.Data.Data, "", "  ")
	l.Debug(string(stats))
	l.Debug(result.Info)
}

func TestClient_GetMessageStatusByMessageId(t *testing.T) {
	result, err := client.GetMessageStatusByMessageId("scm55282525064870278fz")
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data)
	l.Debug(result.Description)
}

func TestClient_GetMessageStatusByJobKey(t *testing.T) {
	result, err := client.GetMessageStatusByJobKey("job_1")
	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data)
	l.Debug(result.Description)
}

func TestClient_GetMessageStatusByRange(t *testing.T) {
	start := (time.Now().UnixNano() - int64(time.Hour * 1)) / int64(time.Millisecond)
	end := time.Now().UnixNano() / int64(time.Millisecond)

	result, err := client.GetMessageStatusByRange(start, end)

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data)
	l.Debug(result.Description)
}

func TestClient_SubscribeForRegId(t *testing.T) {
	result, err := client.SubscribeForRegId(&regId, "topic_test_1", "")

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_UnsubscribeForRegId(t *testing.T) {
	result, err := client.UnsubscribeForRegId(&regId, "topic_test_1", "")

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_SubscribeForAlias(t *testing.T) {
	result, err := client.SubscribeForAlias(&alias, "topic_test_1", "")

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_UnsubscribeForAlias(t *testing.T) {
	result, err := client.UnsubscribeForAlias(&alias, "topic_test_1", "")

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_FetchInvalidRegIds(t *testing.T) {
	result, err := client.FetchInvalidRegIds()

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.List)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_GetRegIdAlias(t *testing.T) {
	result, err := client.GetRegIdAlias(regId[0])

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.List)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_GetRegIdTopic(t *testing.T) {
	result, err := client.GetRegIdTopic(regId[0])

	if err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 {
		t.Fatal(result.Code, result.Description)
	}

	l.Debug(result.Data.List)
	l.Debug(result.Code, " ", result.Description)
}

func TestClient_ScheduleJobExist(t *testing.T) {
	result, err := client.ScheduleJobExist("scm55282525064870278fz")

	if err != nil {
		t.Fatal(err)
	}

	l.Debug(result)
}

func TestClient_ScheduleJobDelete(t *testing.T) {
	result, err := client.ScheduleJobDelete("scm55282525064870278fz")

	if err != nil {
		t.Fatal(err)
	}

	l.Debug(result)
}
