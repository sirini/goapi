package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BadgeRepository interface {
	Award(param models.BadgeAwardParam) (bool, error)
	CreateDefinition(definition models.BadgeDefinition) error
	FindForUser(userUid uint, inlineOnly bool) ([]models.UserBadge, error)
	FindFeaturedForUsers(userUids []uint) (map[uint][]models.UserBadge, error)
	FindUnannouncedForUser(userUid uint, limit uint) ([]models.UserBadge, error)
	ListDefinitions() ([]models.BadgeDefinition, error)
	MarkAnnounced(userUid uint, badgeKeys []string, announcedAt uint64) error
	RecordPostOrigin(param models.PostOriginParam) error
	UpdateDefinition(definition models.BadgeDefinition) (bool, error)
}

func (r *NuboBadgeRepository) FindUnannouncedForUser(userUid uint, limit uint) ([]models.UserBadge, error) {
	badges := make([]models.UserBadge, 0)
	query := fmt.Sprintf(`SELECT d.badge_key, d.name, d.description, d.icon_key, ub.qualified_at
		FROM %s%s AS ub JOIN %s%s AS d ON d.badge_key = ub.badge_key
		WHERE ub.user_uid = ? AND ub.announced_at = 0 AND d.active = 1
		ORDER BY ub.awarded_at ASC, d.sort_order ASC, d.badge_key ASC LIMIT ?`,
		configs.Env.Prefix, models.TABLE_USER_BADGE, configs.Env.Prefix, models.TABLE_BADGE)
	rows, err := r.db.Query(query, userUid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		badge := models.UserBadge{}
		if err := rows.Scan(&badge.Key, &badge.Name, &badge.Description, &badge.IconKey, &badge.EarnedAt); err != nil {
			return nil, err
		}
		badges = append(badges, badge)
	}
	return badges, rows.Err()
}

func (r *NuboBadgeRepository) MarkAnnounced(userUid uint, badgeKeys []string, announcedAt uint64) error {
	if len(badgeKeys) == 0 {
		return nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(badgeKeys)), ",")
	args := make([]any, 0, len(badgeKeys)+2)
	args = append(args, announcedAt, userUid)
	for _, key := range badgeKeys {
		args = append(args, key)
	}
	query := fmt.Sprintf(`UPDATE %s%s SET announced_at = ?
		WHERE user_uid = ? AND announced_at = 0 AND badge_key IN (%s)`,
		configs.Env.Prefix, models.TABLE_USER_BADGE, placeholders)
	_, err := r.db.Exec(query, args...)
	return err
}

type NuboBadgeRepository struct {
	db *sql.DB
}

func NewNuboBadgeRepository(db *sql.DB) *NuboBadgeRepository {
	return &NuboBadgeRepository{db: db}
}

func (r *NuboBadgeRepository) Award(param models.BadgeAwardParam) (bool, error) {
	query := fmt.Sprintf(`INSERT IGNORE INTO %s%s
		(user_uid, badge_key, qualified_at, awarded_at, grant_source, granted_by, evidence_type, evidence_uid)
		SELECT ?, badge_key, ?, ?, ?, ?, ?, ? FROM %s%s WHERE badge_key = ? AND active = 1`,
		configs.Env.Prefix, models.TABLE_USER_BADGE, configs.Env.Prefix, models.TABLE_BADGE)
	result, err := r.db.Exec(query,
		param.UserUid, param.QualifiedAt, time.Now().UnixMilli(), param.GrantSource, param.GrantedBy,
		param.EvidenceType, param.EvidenceUid, param.BadgeKey,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (r *NuboBadgeRepository) CreateDefinition(definition models.BadgeDefinition) error {
	query := fmt.Sprintf(`INSERT INTO %s%s
		(badge_key, name, description, icon_key, rule_key, active, show_inline, sort_order, created, updated)
		VALUES (?, ?, ?, ?, '', 1, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_BADGE)
	_, err := r.db.Exec(query, definition.Key, definition.Name, definition.Description, definition.IconKey,
		definition.ShowInline, definition.SortOrder, definition.Created, definition.Updated)
	return err
}

func (r *NuboBadgeRepository) ListDefinitions() ([]models.BadgeDefinition, error) {
	definitions := make([]models.BadgeDefinition, 0)
	query := fmt.Sprintf(`SELECT badge_key, name, description, icon_key, active, show_inline,
		sort_order, rule_key <> '', created, updated FROM %s%s
		ORDER BY sort_order ASC, created ASC, badge_key ASC`, configs.Env.Prefix, models.TABLE_BADGE)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		definition := models.BadgeDefinition{}
		if err := rows.Scan(&definition.Key, &definition.Name, &definition.Description, &definition.IconKey,
			&definition.Active, &definition.ShowInline, &definition.SortOrder, &definition.System,
			&definition.Created, &definition.Updated); err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	return definitions, rows.Err()
}

func (r *NuboBadgeRepository) UpdateDefinition(definition models.BadgeDefinition) (bool, error) {
	query := fmt.Sprintf(`UPDATE %s%s SET name = ?, description = ?, icon_key = ?,
		show_inline = ?, sort_order = ?, updated = ? WHERE badge_key = ? AND rule_key = '' LIMIT 1`,
		configs.Env.Prefix, models.TABLE_BADGE)
	result, err := r.db.Exec(query, definition.Name, definition.Description, definition.IconKey,
		definition.ShowInline, definition.SortOrder, definition.Updated, definition.Key)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (r *NuboBadgeRepository) FindForUser(userUid uint, inlineOnly bool) ([]models.UserBadge, error) {
	badges := make([]models.UserBadge, 0)
	inlineClause := ""
	if inlineOnly {
		inlineClause = " AND d.show_inline = 1"
	}
	query := fmt.Sprintf(`SELECT d.badge_key, d.name, d.description, d.icon_key, ub.qualified_at
		FROM %s%s AS ub JOIN %s%s AS d ON d.badge_key = ub.badge_key
		WHERE ub.user_uid = ? AND d.active = 1%s
		ORDER BY d.sort_order ASC, ub.qualified_at ASC, d.badge_key ASC`,
		configs.Env.Prefix, models.TABLE_USER_BADGE,
		configs.Env.Prefix, models.TABLE_BADGE,
		inlineClause,
	)
	rows, err := r.db.Query(query, userUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		badge := models.UserBadge{}
		if err := rows.Scan(&badge.Key, &badge.Name, &badge.Description, &badge.IconKey, &badge.EarnedAt); err != nil {
			return nil, err
		}
		badges = append(badges, badge)
	}
	return badges, rows.Err()
}

func (r *NuboBadgeRepository) FindFeaturedForUsers(userUids []uint) (map[uint][]models.UserBadge, error) {
	result := make(map[uint][]models.UserBadge)
	if len(userUids) == 0 {
		return result, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(userUids)), ",")
	args := make([]any, len(userUids))
	for i, uid := range userUids {
		args[i] = uid
	}
	query := fmt.Sprintf(`SELECT ub.user_uid, d.badge_key, d.name, d.description, d.icon_key, ub.qualified_at
		FROM %s%s AS ub JOIN %s%s AS d ON d.badge_key = ub.badge_key
		WHERE d.active = 1 AND d.show_inline = 1 AND ub.user_uid IN (%s)
		ORDER BY d.sort_order ASC, ub.qualified_at ASC, d.badge_key ASC`,
		configs.Env.Prefix, models.TABLE_USER_BADGE,
		configs.Env.Prefix, models.TABLE_BADGE,
		placeholders,
	)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userUid uint
		badge := models.UserBadge{}
		if err := rows.Scan(&userUid, &badge.Key, &badge.Name, &badge.Description, &badge.IconKey, &badge.EarnedAt); err != nil {
			return nil, err
		}
		result[userUid] = append(result[userUid], badge)
	}
	return result, rows.Err()
}

func (r *NuboBadgeRepository) RecordPostOrigin(param models.PostOriginParam) error {
	query := fmt.Sprintf(`INSERT INTO %s%s (post_uid, client_key, app_version, recorded_at)
		VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE client_key = VALUES(client_key),
		app_version = VALUES(app_version), recorded_at = VALUES(recorded_at)`,
		configs.Env.Prefix, models.TABLE_POST_ORIGIN)
	_, err := r.db.Exec(query, param.PostUid, param.ClientKey, param.AppVersion, param.RecordedAt)
	return err
}
