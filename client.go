/*
小米推送 golang sdk

快速开始

	// 创建 client， 支持多个 package
	client := xmpush.NewClient("appSecret", "packageName")

	message := xmpush.NewMessage("title", "description")
	message.SetNotifyType(1)

	result, err := client.SendToRegId(message, "regId")
	if err != nil {
		// handle error
	}

	// handle result
*/
package xmpush

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	sandbox    = "https://sandbox.xmpush.xiaomi.com"
	production = "https://api.xmpush.xiaomi.com"

	apiRetryTimes = 3

	regIdURL   = "/v3/message/regid"
	aliasURL   = "/v3/message/alias"
	accountURL = "/v2/message/user_account"
	allURL     = "/v3/message/all"
	TopicURL   = "/v3/message/topic"
	TopicOpURL = "/v3/message/multi_topic"

	multiRegIdURL   = "/v2/multi_messages/regids"
	multiAliasURL   = "/v2/multi_messages/aliases"
	multiAccountURL = "/v2/multi_messages/user_accounts"

	statsURL          = "/v1/stats/message/counters"
	messageStatusURL  = "/v1/trace/message/status"
	messagesStatusURL = "/v1/trace/messages/status"

	subscribeURL   = "/v2/topic/subscribe"
	unsubscribeURL = "/v2/topic/unsubscribe"

	subscribeAliasURL   = "/v2/topic/subscribe/alias"
	unsubscribeAliasURL = "/v2/topic/unsubscribe/alias"

	invalidRegIdsURL = "https://feedback.xmpush.xiaomi.com/v1/feedback/fetch_invalid_regids"

	regIdAliasURL = "/v1/alias/all"
	regIdTopicURL = "/v1/topic/all"

	scheduleJobExistURL  = "/v2/schedule_job/exist"
	scheduleJobDeleteURL = "/v2/schedule_job/delete"
)

// 创建客户端
func NewClient(appSecret string, packageNames ...string) (*Client, error) {
	if appSecret == "" || packageNames == nil || len(packageNames) == 0 {
		return nil, errors.New("error params")
	}
	return &Client{
		useSandbox:          false,
		appSecret:           appSecret,
		packageNames:        packageNames,
		hasMultiPackageName: len(packageNames) > 1,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		log: &nopeLogger{},
	}, nil
}

type Client struct {
	useSandbox          bool
	appSecret           string
	packageNames        []string
	hasMultiPackageName bool
	client              *http.Client
	log                 logger
}

func (c *Client) UseSandbox(use bool) {
	c.useSandbox = use
}

func (c *Client) SetLogger(logger logger) {
	c.log = logger
}

// 向 regId 发送单条消息
func (c *Client) SendToRegId(message *Message, regId *[]string) (*SendResult, error) {
	param, err := c.buildParam(message, "registration_id", regId)
	if err != nil {
		return nil, err
	}

	res, err := c.doPost(regIdURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 向 alias 发送单条消息
func (c *Client) SendToAlias(message *Message, alias *[]string) (*SendResult, error) {
	param, err := c.buildParam(message, "alias", alias)
	if err != nil {
		return nil, err
	}

	res, err := c.doPost(aliasURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 向 account 发送单条消息
func (c *Client) SendToAccount(message *Message, account *[]string) (*SendResult, error) {
	param, err := c.buildParam(message, "user_account", account)
	if err != nil {
		return nil, err
	}

	res, err := c.doPost(accountURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 向 topic 发送单条消息
func (c *Client) SendToTopic(message *Message, topic string) (*SendResult, error) {
	param, err := c.messageToForm(message)
	if err != nil {
		return nil, err
	}

	param.Add("topic", topic)

	res, err := c.doPost(TopicURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type TopicOP string

const (
	TopicOPUnion        TopicOP = "UNION"        // 并集
	TopicOPIntersection TopicOP = "INTERSECTION" // 交集
	TopicOPExcept       TopicOP = "EXCEPT"       // 差集
)

// 向 多个 topic 发送单条消息
//
// topics 为 2 ~ 5 个， topicOP 为空时，默认为取并集
func (c *Client) SendToTopics(message *Message, topics *[]string, topicOP TopicOP) (*SendResult, error) {
	param, err := c.messageToForm(message)
	if err != nil {
		return nil, err
	}

	if topics == nil {
		return nil, errors.New("topics should not nil")
	}

	ts := *topics
	tc := len(ts)

	if tc < 2 || tc > 5 {
		return nil, errors.New("topics count should between 1 and 5")
	}

	if topicOP == "" {
		topicOP = TopicOPUnion
	}

	if topicOP != TopicOPUnion &&
		topicOP != TopicOPIntersection &&
		topicOP != TopicOPExcept {
		return nil, errors.New("unknown topic option")
	}

	param.Add("topics", strings.Join(ts, ";$;"))
	param.Add("topic_op", string(topicOP))

	res, err := c.doPost(TopicOpURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 推送多条消息 (regId, alias, account) 通过 targetType 判断
// 一次只支持一种 targetType
func (c *Client) SendTargetedMessage(messages *[]TargetedMessage) (*SendResult, error) {
	if messages == nil || len(*messages) == 0 {
		return nil, errors.New("messages can't empty")
	}

	param, err := c.buildTargetedMessageParam(messages)
	if err != nil {
		return nil, err
	}

	var apiURI string
	targetType := (*messages)[0].targetType
	switch targetType {
	case TargetRegId:
		apiURI = multiRegIdURL
	case TargetAlias:
		apiURI = multiAliasURL
	case TargetAccount:
		apiURI = multiAccountURL
	default:
		return nil, fmt.Errorf("unknown target type %d", targetType)
	}

	res, err := c.doPost(apiURI, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 向 所有设备 发送单条消息
func (c *Client) SendToAll(message *Message) (*SendResult, error) {
	param, err := c.messageToForm(message)
	if err != nil {
		return nil, err
	}

	res, err := c.doPost(allURL, param)
	if err != nil {
		return nil, err
	}

	var result SendResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 获取消息的统计数据
// start, end 格式为： yyyyMMdd
func (c *Client) Stats(start, end string) (*StatsResult, error) {
	form := &url.Values{}
	form.Add("start_date", start)
	form.Add("end_date", end)
	form.Add("restricted_package_name", c.packageNames[0])

	res, err := c.doGet(statsURL, form)
	if err != nil {
		return nil, err
	}

	var result StatsResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 追踪消息状态 - messageId
func (c *Client) GetMessageStatusByMessageId(messageId string) (*SingleStatusResult, error) {
	if messageId == "" {
		return nil, errors.New("message id can't empty")
	}

	form := &url.Values{}
	form.Add("msg_id", messageId)

	res, err := c.doGet(messageStatusURL, form)
	if err != nil {
		return nil, err
	}

	var result SingleStatusResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 追踪消息状态 - jobKey
func (c *Client) GetMessageStatusByJobKey(jobKey string) (*SingleStatusResult, error) {
	if jobKey == "" {
		return nil, errors.New("jobKey can't empty")
	}

	form := &url.Values{}
	form.Add("job_key", jobKey)

	res, err := c.doGet(messageStatusURL, form)
	if err != nil {
		return nil, err
	}

	var result SingleStatusResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 追踪消息状态 - time range
func (c *Client) GetMessageStatusByRange(beginTimestamp, endTimestamp int64) (*BatchStatusResult, error) {
	form := &url.Values{}
	form.Add("begin_time", fmt.Sprintf("%d", beginTimestamp))
	form.Add("end_time", fmt.Sprintf("%d", endTimestamp))

	res, err := c.doGet(messagesStatusURL, form)
	if err != nil {
		return nil, err
	}

	var result BatchStatusResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 订阅标签 regId
func (c *Client) SubscribeForRegId(regId *[]string, topic string, category string) (*Result, error) {
	return c.subscribeAction("regId", subscribeURL, regId, topic, category)
}

// 取消订阅标签 regId
func (c *Client) UnsubscribeForRegId(regId *[]string, topic string, category string) (*Result, error) {
	return c.subscribeAction("regId", unsubscribeURL, regId, topic, category)
}

// 订阅标签 alias
func (c *Client) SubscribeForAlias(regId *[]string, topic string, category string) (*Result, error) {
	return c.subscribeAction("alias", subscribeAliasURL, regId, topic, category)
}

// 取消订阅标签 alias
func (c *Client) UnsubscribeForAlias(regId *[]string, topic string, category string) (*Result, error) {
	return c.subscribeAction("alias", unsubscribeAliasURL, regId, topic, category)
}

func (c *Client) subscribeAction(actionType string, actionURI string, targets *[]string,
	topic string, category string) (*Result, error) {
	form := &url.Values{}

	var key string
	switch actionType {
	case "regId":
		key = "registration_id"
	case "alias":
		key = "aliases"
	default:
		return nil, errors.New("unknown actionType")
	}
	form.Add(key, strings.Join(*targets, ","))

	form.Add("topic", topic)
	if category != "" {
		form.Add("category", category)
	}

	if c.hasMultiPackageName {
		form.Add("restricted_package_name", strings.Join(c.packageNames, ","))
	}

	res, err := c.doPost(actionURI, form)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 获取失效的regId
func (c *Client) FetchInvalidRegIds() (*InvalidRegIdsResult, error) {
	res, err := c.doGet(invalidRegIdsURL, nil)
	if err != nil {
		return nil, err
	}

	var result InvalidRegIdsResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 用户目前设置的所有Alias
func (c *Client) GetRegIdAlias(regId string) (*RegIdAalisResult, error) {
	form := &url.Values{}
	form.Add("registration_id", regId)
	if c.hasMultiPackageName {
		form.Add("restricted_package_name", strings.Join(c.packageNames, ","))
	}

	res, err := c.doGet(regIdAliasURL, form)
	if err != nil {
		return nil, err
	}

	var result RegIdAalisResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 用户目前订阅的所有Topic
func (c *Client) GetRegIdTopic(regId string) (*RegIdTopicResult, error) {
	form := &url.Values{}
	form.Add("registration_id", regId)
	if c.hasMultiPackageName {
		form.Add("restricted_package_name", strings.Join(c.packageNames, ","))
	}

	res, err := c.doGet(regIdTopicURL, form)
	if err != nil {
		return nil, err
	}

	var result RegIdTopicResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 定时任务是否存在
// result.code = 0 为任务存在, 否则不存在
func (c *Client) ScheduleJobExist(messageId string) (*Result, error) {
	form := &url.Values{}
	form.Add("job_id", messageId)

	res, err := c.doPost(scheduleJobExistURL, form)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 删除定时任务
func (c *Client) ScheduleJobDelete(messageId string) (*Result, error) {
	form := &url.Values{}
	form.Add("job_id", messageId)

	res, err := c.doPost(scheduleJobDeleteURL, form)
	if err != nil {
		return nil, err
	}

	var result Result
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) doPost(api string, form *url.Values) ([]byte, error) {
	param := ""
	if form != nil {
		param = form.Encode()
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s", c.buildURI(api)),
		strings.NewReader(param))

	if err != nil {
		return nil, err
	}

	c.log.Debugf("doPost request url: %v, body: %v", req.URL, form.Encode())

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	return c.doReq(req)
}

func (c *Client) doGet(api string, form *url.Values) ([]byte, error) {
	param := ""
	if form != nil {
		param = form.Encode()
	}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s?%s", c.buildURI(api), param),
		nil)

	if err != nil {
		return nil, err
	}

	c.log.Debugf("doGet request url: %v", req.URL)

	return c.doReq(req)
}

func (c *Client) doReq(req *http.Request) ([]byte, error) {
	req.Header.Add("Authorization", fmt.Sprintf("key=%s", c.appSecret))

	var lastError error
	var body []byte
	for i := 0; i < apiRetryTimes; i += 1 {
		lastError = nil
		res, err := c.client.Do(req)
		lastError = err

		if err != nil {
			c.log.Debug("error", err)
		}

		if err == nil {
			if res.Body == nil {
				lastError = errors.New("xiaomi push API response nil")
				continue
			}

			c.log.Debugf("request time %d, status: %s", i, res.Status)
			body, err = ioutil.ReadAll(res.Body)
			res.Body.Close()

			if err != nil {
				lastError = err
				continue
			}

			if res.StatusCode != http.StatusOK {
				lastError = fmt.Errorf("xiaomi push API status %d, response %v", res.StatusCode, string(body))
				continue
			}

			break
		}
	}

	c.log.Debug("response: ", string(body))

	return body, lastError
}

func (c *Client) buildURI(uri string) string {
	if strings.HasPrefix(uri, "http") {
		return uri
	}
	return fmt.Sprintf("%s%s", c.baseURI(), uri)
}

func (c *Client) baseURI() string {
	if c.useSandbox {
		return sandbox
	}
	return production
}

func (c *Client) validateMessage(message *Message) (bool, error) {
	if message.Title == "" || message.Description == "" {
		return false, errors.New("title and description can't empty")
	}

	notifyType := message.NotifyType
	if notifyType != -1 &&
		notifyType != 1 &&
		notifyType != 2 &&
		notifyType != 4 {
		return false, fmt.Errorf("unknown notifyType %d", notifyType)
	}

	if message.TimeToLive > 0 {
		t := time.Now().Add(time.Duration(message.TimeToLive) * time.Millisecond)
		if t.After(time.Now().Add(time.Duration(MaxTimeToLive) * time.Millisecond)) {
			return false, fmt.Errorf("TimeToLive too long (%s %v)", message.Title, message.TimeToLive)
		}
	}

	if message.TimeToSend > 0 {
		t := time.Now().Add(MaxTimeToSend)
		sc := time.Unix(0, message.TimeToSend*int64(time.Millisecond))
		if t.Before(sc) {
			return false, fmt.Errorf("TimeToSend error (%s %v), should before %v",
				message.Title, message.TimeToSend, t.Format(time.RFC3339))
		}

		if sc.Before(time.Now()) {
			return false, fmt.Errorf("TimeToSend error (%s %v), should after (now) %v",
				message.Title, message.TimeToSend, time.Now().Format(time.RFC3339))
		}
	}

	return true, nil
}

func (c *Client) messageToForm(message *Message) (*url.Values, error) {
	if _, err := c.validateMessage(message); err != nil {
		return nil, err
	}

	form := &url.Values{}

	form.Add("restricted_package_name", strings.Join(c.packageNames, ","))
	form.Add("payload", message.Payload)
	form.Add("title", message.Title)
	form.Add("description", message.Description)
	form.Add("notify_type", fmt.Sprintf("%d", message.NotifyType))
	form.Add("pass_through", fmt.Sprintf("%d", message.PassThrough))

	if message.NotifyID > 0 {
		form.Add("notify_id", fmt.Sprintf("%d", message.NotifyID))
	}

	if message.TimeToLive > 0 {
		form.Add("time_to_live", fmt.Sprintf("%d", message.TimeToLive))
	}

	if message.TimeToSend > 0 {
		form.Add("time_to_send", fmt.Sprintf("%d", message.TimeToSend))
	}

	if message.Extra != nil {
		for k, v := range message.Extra {
			form.Add(fmt.Sprintf("extra.%s", k), v)
		}
	}

	return form, nil
}

func (c *Client) buildParam(message *Message, paramType string, item *[]string) (*url.Values, error) {
	form, err := c.messageToForm(message)
	if err != nil {
		return nil, err
	}

	err = errors.New("count should more than 1 and less than 1000")

	if item == nil {
		return nil, err
	}

	count := len(*item)
	if count == 0 || count > 1000 {
		return nil, err
	}

	form.Add(paramType, strings.Join(*item, ","))
	return form, nil
}

func (c *Client) buildTargetedMessageParam(messages *[]TargetedMessage) (*url.Values, error) {
	type M struct {
		Target  string   `json:"target"`
		Message *Message `json:"message"`
	}

	var ms []M
	for _, m := range *messages {
		if _, err := c.validateMessage(m.message); err != nil {
			return nil, err
		}

		ms = append(ms, M{
			Target:  m.target,
			Message: m.message,
		})
	}

	data, err := json.Marshal(&ms)
	if err != nil {
		return nil, err
	}

	form := &url.Values{}
	jsonMessages := string(data)
	c.log.Debug("target messages json: ", jsonMessages)
	form.Add("messages", jsonMessages)

	return form, nil
}
