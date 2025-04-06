package lib

// MyConfig описывает структурку, получаемую при загрузке и распарсивании конфига
type MyConfig struct {
	Server      string `json:"server,omitempty"`
	Port        int    `json:"port,omitempty"`
	LogLevel    string `json:"loglevel,omitempty"`
	Log         string `json:"log,omitempty"`
	Channel     string `json:"channel,omitempty"`
	DataDir     string `json:"data_dir,omitempty"`
	CSign       string `json:"csign,omitempty"`
	ForwardsMax int64  `json:"forwards_max,omitempty"`
}

// Входящее сообщение из pubsub-канала redis-ки
type rMsg struct {
	From     string `json:"from,omitempty"`
	ChatID   string `json:"chatid,omitempty"`
	UserID   string `json:"userid,omitempty"`
	ThreadID string `json:"threadid"`
	Message  string `json:"message,omitempty"`
	Plugin   string `json:"plugin,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Misc     struct {
		Answer      int64  `json:"answer,omitempty"`
		BotNick     string `json:"bot_nick,omitempty"`
		CSign       string `json:"csign,omitempty"`
		FwdCnt      int64  `json:"fwd_cnt,omitempty"`
		GoodMorning int64  `json:"good_morning,omitempty"`
		MsgFormat   int64  `json:"msg_format,omitempty"`
		Username    string `json:"username,omitempty"`
	} `json:"Misc"`
}

// Исходящее сообщение в pubsub-канал redis-ки
type sMsg struct {
	From     string `json:"from"`
	ChatID   string `json:"chatid"`
	Userid   string `json:"userid"`
	ThreadID string `json:"threadid"`
	Message  string `json:"message"`
	Plugin   string `json:"plugin"`
	Mode     string `json:"mode"`
	Misc     struct {
		Answer      int64  `json:"answer"`
		BotNick     string `json:"bot_nick"`
		CSign       string `json:"csign"`
		FwdCnt      int64  `json:"fwd_cnt"`
		GoodMorning int64  `json:"good_morning"`
		MsgFormat   int64  `json:"msg_format"`
		Username    string `json:"username"`
	} `json:"misc"`
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
