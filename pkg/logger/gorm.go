package logger

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`
	Level     string    `gorm:"size:20;index" json:"level"`
	TraceID   string    `gorm:"size:64;index" json:"trace_id"`
	UserID    string    `gorm:"size:20;index" json:"user_id"`
	Tag       string    `gorm:"size:32;index" json:"tag"`
	Message   string    `gorm:"size:1024" json:"message"`
	Stack     string    `gorm:"type:text" json:"stack"`
	Data      string    `gorm:"type:text" json:"data"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

type GormHook struct {
	db *gorm.DB
}

func NewGormHook(db *gorm.DB) *GormHook {
	err := db.AutoMigrate(new(Logger))
	if err != nil {
		panic(err)
	}

	return &GormHook{
		db: db,
	}
}

func (h *GormHook) Exec(extra map[string]string, b []byte) error {
	msg := &Logger{
		ID: xid.New().String(),
	}
	data := make(map[string]interface{})
	err := jsoniter.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	if v, ok := data["ts"]; ok {
		msg.CreatedAt = time.UnixMilli(int64(v.(float64)))
		delete(data, "ts")
	}
	if v, ok := data["msg"]; ok {
		msg.Message = v.(string)
		delete(data, "msg")
	}
	if v, ok := data["tag"]; ok {
		msg.Tag = v.(string)
		delete(data, "tag")
	}
	if v, ok := data["trace_id"]; ok {
		msg.TraceID = v.(string)
		delete(data, "trace_id")
	}
	if v, ok := data["user_id"]; ok {
		msg.UserID = v.(string)
		delete(data, "user_id")
	}
	if v, ok := data["level"]; ok {
		msg.Level = v.(string)
		delete(data, "level")
	}
	if v, ok := data["stack"]; ok {
		msg.Stack = v.(string)
		delete(data, "stack")
	}
	delete(data, "caller")

	for k, v := range extra {
		data[k] = v
	}
	if len(data) > 0 {
		buf, _ := jsoniter.Marshal(data)
		msg.Data = string(buf)
	}

	return h.db.Create(msg).Error
}

func (h *GormHook) Close() error {
	db, err := h.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
