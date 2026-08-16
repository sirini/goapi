package repositories

import (
	"database/sql"
	"fmt"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type MailDeliveryRepository interface {
	CreateDelivery(record models.MailDeliveryRecord) error
	ListDeliveries(param models.MailDeliveryListParam, since uint64) (models.MailDeliveryListResult, error)
}

type NuboMailDeliveryRepository struct {
	db *sql.DB
}

func NewNuboMailDeliveryRepository(db *sql.DB) *NuboMailDeliveryRepository {
	return &NuboMailDeliveryRepository{db: db}
}

func (r *NuboMailDeliveryRepository) CreateDelivery(record models.MailDeliveryRecord) error {
	query := fmt.Sprintf(`INSERT INTO %s%s
		(type, recipient, subject, provider, provider_message_id, status, error, created)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_MAIL_DELIVERY)
	_, err := r.db.Exec(query, record.Type, record.Recipient, record.Subject, record.Provider,
		record.ProviderMessageID, record.Status, record.Error, record.Created)
	return err
}

func (r *NuboMailDeliveryRepository) ListDeliveries(param models.MailDeliveryListParam, since uint64) (models.MailDeliveryListResult, error) {
	if param.Page < 1 {
		param.Page = 1
	}
	if param.Limit < 1 || param.Limit > 100 {
		param.Limit = 20
	}
	result := models.MailDeliveryListResult{
		Items: make([]models.MailDeliveryRecord, 0),
		Page:  param.Page,
		Limit: param.Limit,
		Summary: models.MailDeliverySummary{
			Since: since,
		},
	}
	table := configs.Env.Prefix + string(models.TABLE_MAIL_DELIVERY)
	if err := r.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&result.Total); err != nil {
		return result, err
	}
	query := fmt.Sprintf(`SELECT uid, type, recipient, subject, provider, provider_message_id, status, error, created
		FROM %s ORDER BY uid DESC LIMIT ? OFFSET ?`, table)
	rows, err := r.db.Query(query, param.Limit, (param.Page-1)*param.Limit)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.MailDeliveryRecord{}
		if err := rows.Scan(&item.Uid, &item.Type, &item.Recipient, &item.Subject, &item.Provider,
			&item.ProviderMessageID, &item.Status, &item.Error, &item.Created); err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	summaryQuery := fmt.Sprintf(`SELECT
		COALESCE(SUM(status = ?), 0),
		COALESCE(SUM(status = ?), 0),
		COALESCE(SUM(status = ? AND type = 'signup-verification'), 0),
		COALESCE(SUM(status = ? AND type = 'password-reset'), 0),
		COALESCE(SUM(status = ? AND type = 'comment-notification'), 0)
		FROM %s WHERE created >= ?`, table)
	err = r.db.QueryRow(summaryQuery,
		models.MailDeliveryAccepted,
		models.MailDeliveryFailed,
		models.MailDeliveryAccepted,
		models.MailDeliveryAccepted,
		models.MailDeliveryAccepted,
		since,
	).Scan(
		&result.Summary.Accepted,
		&result.Summary.Failed,
		&result.Summary.SignupVerification,
		&result.Summary.PasswordReset,
		&result.Summary.CommentNotification,
	)
	return result, err
}
