package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type MailCampaignRepository interface {
	CreateCampaign(param models.MailCampaignSaveParam) (uint, error)
	UpdateCampaign(param models.MailCampaignSaveParam) error
	GetCampaign(uid uint) (models.MailCampaign, error)
	ListCampaigns(limit uint) (models.MailCampaignListResult, error)
	GetActiveMailRecipients() ([]models.MailRecipient, error)
	SetCampaignImport(uid uint, segmentId, importId string, recipientCount uint) error
	SetCampaignStatus(uid uint, status, lastError string) error
	BeginCampaignSend(uid uint) error
	SetCampaignBroadcast(uid uint, broadcastId string) error
	SetCampaignSent(uid uint, broadcastId string) error
}

type NuboMailCampaignRepository struct {
	db *sql.DB
}

func NewNuboMailCampaignRepository(db *sql.DB) *NuboMailCampaignRepository {
	return &NuboMailCampaignRepository{db: db}
}

func (r *NuboMailCampaignRepository) CreateCampaign(param models.MailCampaignSaveParam) (uint, error) {
	now := time.Now().UnixMilli()
	query := fmt.Sprintf(`INSERT INTO %s%s (subject, markdown, status, created, updated)
		VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, param.Subject, param.Markdown, models.MailCampaignDraft, now, now)
	if err != nil {
		return 0, err
	}
	uid, err := result.LastInsertId()
	return uint(uid), err
}

func (r *NuboMailCampaignRepository) UpdateCampaign(param models.MailCampaignSaveParam) error {
	query := fmt.Sprintf(`UPDATE %s%s SET subject = ?, markdown = ?, updated = ?, last_error = ''
		WHERE uid = ? AND status != ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, param.Subject, param.Markdown, time.Now().UnixMilli(), param.Uid, models.MailCampaignSent)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("campaign does not exist or was already sent")
	}
	return nil
}

func (r *NuboMailCampaignRepository) GetCampaign(uid uint) (models.MailCampaign, error) {
	result := models.MailCampaign{}
	query := fmt.Sprintf(`SELECT uid, subject, markdown, status, recipient_count, resend_segment_id,
		resend_import_id, resend_broadcast_id, last_error, created, updated, sent
		FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	err := r.db.QueryRow(query, uid).Scan(&result.Uid, &result.Subject, &result.Markdown, &result.Status,
		&result.RecipientCount, &result.ResendSegmentId, &result.ResendImportId, &result.ResendBroadcastId,
		&result.LastError, &result.Created, &result.Updated, &result.Sent)
	return result, err
}

func (r *NuboMailCampaignRepository) ListCampaigns(limit uint) (models.MailCampaignListResult, error) {
	result := models.MailCampaignListResult{Items: make([]models.MailCampaign, 0)}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	table := configs.Env.Prefix + string(models.TABLE_MAIL_CAMPAIGN)
	if err := r.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&result.Total); err != nil {
		return result, err
	}
	query := fmt.Sprintf(`SELECT uid, subject, markdown, status, recipient_count, resend_segment_id,
		resend_import_id, resend_broadcast_id, last_error, created, updated, sent
		FROM %s ORDER BY uid DESC LIMIT ?`, table)
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.MailCampaign{}
		if err := rows.Scan(&item.Uid, &item.Subject, &item.Markdown, &item.Status, &item.RecipientCount,
			&item.ResendSegmentId, &item.ResendImportId, &item.ResendBroadcastId, &item.LastError,
			&item.Created, &item.Updated, &item.Sent); err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func (r *NuboMailCampaignRepository) GetActiveMailRecipients() ([]models.MailRecipient, error) {
	result := make([]models.MailRecipient, 0)
	query := fmt.Sprintf(`SELECT id, name FROM %s%s WHERE id != '' AND blocked = 0 ORDER BY uid`,
		configs.Env.Prefix, models.TABLE_USER)
	rows, err := r.db.Query(query)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		item := models.MailRecipient{}
		if err := rows.Scan(&item.Email, &item.Name); err != nil {
			return result, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *NuboMailCampaignRepository) SetCampaignImport(uid uint, segmentId, importId string, recipientCount uint) error {
	query := fmt.Sprintf(`UPDATE %s%s SET status = ?, resend_segment_id = ?, resend_import_id = ?,
		recipient_count = ?, last_error = '', updated = ? WHERE uid = ? AND status != ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, models.MailCampaignSyncing, segmentId, importId, recipientCount,
		time.Now().UnixMilli(), uid, models.MailCampaignSent)
	return requireOneCampaignRow(result, err)
}

func (r *NuboMailCampaignRepository) SetCampaignStatus(uid uint, status, lastError string) error {
	query := fmt.Sprintf(`UPDATE %s%s SET status = ?, last_error = ?, updated = ?
		WHERE uid = ? AND status != ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, status, lastError, time.Now().UnixMilli(), uid, models.MailCampaignSent)
	return requireOneCampaignRow(result, err)
}

func (r *NuboMailCampaignRepository) SetCampaignBroadcast(uid uint, broadcastId string) error {
	query := fmt.Sprintf(`UPDATE %s%s SET resend_broadcast_id = ?, last_error = '', updated = ?
		WHERE uid = ? AND status = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, broadcastId, time.Now().UnixMilli(), uid, models.MailCampaignSending)
	return requireOneCampaignRow(result, err)
}

func (r *NuboMailCampaignRepository) BeginCampaignSend(uid uint) error {
	query := fmt.Sprintf(`UPDATE %s%s SET status = ?, last_error = '', updated = ?
		WHERE uid = ? AND status = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, models.MailCampaignSending, time.Now().UnixMilli(), uid, models.MailCampaignReady)
	return requireOneCampaignRow(result, err)
}

func (r *NuboMailCampaignRepository) SetCampaignSent(uid uint, broadcastId string) error {
	now := time.Now().UnixMilli()
	query := fmt.Sprintf(`UPDATE %s%s SET status = ?, resend_broadcast_id = ?, last_error = '', updated = ?, sent = ?
		WHERE uid = ? AND status = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_MAIL_CAMPAIGN)
	result, err := r.db.Exec(query, models.MailCampaignSent, broadcastId, now, now, uid, models.MailCampaignSending)
	return requireOneCampaignRow(result, err)
}

func requireOneCampaignRow(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("campaign state changed; reload and try again")
	}
	return nil
}
