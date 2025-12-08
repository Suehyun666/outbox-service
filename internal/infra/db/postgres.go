package db

import (
	"database/sql"
	"fmt"
	"outbox-service/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver 등록
)

type PostgresProcessor struct {
	db *sql.DB
}

func NewPostgresProcessor(dsn string) (*PostgresProcessor, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// 커넥션 풀 설정 (중요)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &PostgresProcessor{db: db}, nil
}

// ProcessBatch : 트랜잭션 하나로 "읽기(Lock) -> 전송 -> 삭제" 원자성 보장
func (r *PostgresProcessor) ProcessBatch(batchSize int, tableName string, sendFunc func(domain.OutboxEntry) error) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	// defer로 패닉 시 롤백 안전장치
	defer tx.Rollback()

	// 1. SELECT ... FOR UPDATE SKIP LOCKED (중복 방지 핵심)
	// status = 'PENDING'이고 available_at이 지난 것만
	query := fmt.Sprintf(`
		SELECT id, aggregate_type, aggregate_id, event_type, payload, idempotency_key,
		       status, created_at, available_at
		FROM %s
		WHERE status = 'PENDING' AND available_at <= NOW()
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, tableName)

	rows, err := tx.Query(query, batchSize)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var entries []domain.OutboxEntry
	for rows.Next() {
		var e domain.OutboxEntry
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType,
			&e.Payload, &e.IdempotencyKey, &e.Status, &e.CreatedAt, &e.AvailableAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	rows.Close() // Explicit close before processing

	if len(entries) == 0 {
		return 0, nil // 할 일 없음
	}

	// 2. 외부 전송 (Kafka Send)
	processedIDs := []int64{}
	for _, entry := range entries {
		if err := sendFunc(entry); err != nil {
			// 전송 실패 시 로그 찍고 계속 진행할지, 중단할지는 정책 결정
			// 여기서는 실패한 건 DB에 남겨두고(Rollback 대상 아님, 삭제 리스트에 안 넣음), 성공한 것만 삭제
			fmt.Printf("Failed to send ID %d: %v\n", entry.ID, err)
			continue
		}
		processedIDs = append(processedIDs, entry.ID)
	}

	// 3. 전송 성공한 것만 PUBLISHED로 업데이트
	if len(processedIDs) > 0 {
		// pgx는 ANY($1) 문법 지원
		updateQuery := fmt.Sprintf(`
			UPDATE %s
			SET status = 'PUBLISHED', published_at = NOW()
			WHERE id = ANY($1)`, tableName)
		// Go slice를 Postgres Array로 변환 (pgx 드라이버가 처리)
		if _, err := tx.Exec(updateQuery, processedIDs); err != nil {
			return 0, err
		}
	}

	// 4. 커밋
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(processedIDs), nil
}
