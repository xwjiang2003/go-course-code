package p12

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type OrderEvent struct {
	OrderID   string `json:"order_id"`
	EventType string `json:"event_type"`
	Status    string `json:"status"`
}

var db *sql.DB

func isDuplicate(err error) bool { return false }

func Run(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"127.0.0.1:9092"},
		GroupID:        "order-worker",
		Topic:          "order-events",
		CommitInterval: 0, // 关闭自动提交
	})

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			break
		}
		if err := handle(ctx, m); err != nil {
			continue // 不提交，等重投
		}
		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("commit failed: %v", err)
		}
	}
}

func handle(ctx context.Context, m kafka.Message) error {
	var ev OrderEvent
	if err := json.Unmarshal(m.Value, &ev); err != nil {
		return nil // 毒消息丢弃，避免卡死分区
	}
	bizKey := ev.OrderID + ":" + ev.EventType
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO consumed_msg(biz_key) VALUES (?)`, bizKey); err != nil {
		if isDuplicate(err) {
			return tx.Commit()
		}
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE orders SET status = ? WHERE id = ?`, ev.Status, ev.OrderID); err != nil {
		return err
	}
	return tx.Commit()
}
