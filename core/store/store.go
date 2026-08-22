// Package store 提供业务数据访问层。
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aitrendings/core/model"
	"aitrendings/pkg/db"
)

// New 创建 Store
func New(database *db.Database) *Store {
	return &Store{db: database}
}

// SaveSnapshot 保存每日快照
func (s *Store) SaveSnapshot(ctx context.Context, snap *model.Snapshot) error {
	rawMeta, err := json.Marshal(snap.RawMetadata)
	if err != nil {
		return fmt.Errorf("marshal raw_metadata: %w", err)
	}

	return s.db.Exec(ctx, `
		INSERT INTO trending_snapshots (project_name, category, stars, forks, delta_stars, snapshot_date, raw_metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (project_name, snapshot_date)
		DO UPDATE SET stars = $3, forks = $4, delta_stars = $5, category = $2, raw_metadata = $7
	`, snap.ProjectName, snap.Category, snap.Stars, snap.Forks,
		snap.DeltaStars, snap.SnapshotDate, rawMeta)
}

// GetLatestSnapshot 获取项目最新快照
func (s *Store) GetLatestSnapshot(ctx context.Context, projectName string) (*model.Snapshot, error) {
	row := s.db.QueryRow(ctx, `
		SELECT project_name, category, stars, forks, delta_stars, snapshot_date, raw_metadata
		FROM trending_snapshots
		WHERE project_name = $1
		ORDER BY snapshot_date DESC
		LIMIT 1
	`, projectName)
	return scanSnapshot(row)
}

// GetSnapshotByDate 获取指定日期的快照
func (s *Store) GetSnapshotByDate(ctx context.Context, projectName string, date time.Time) (*model.Snapshot, error) {
	row := s.db.QueryRow(ctx, `
		SELECT project_name, category, stars, forks, delta_stars, snapshot_date, raw_metadata
		FROM trending_snapshots
		WHERE project_name = $1 AND snapshot_date = $2
	`, projectName, date.Format("2006-01-02"))
	return scanSnapshot(row)
}

// GetRecentSnapshots 获取近 N 天所有快照
func (s *Store) GetRecentSnapshots(ctx context.Context, days int) ([]*model.Snapshot, error) {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := s.db.Query(ctx, `
		SELECT project_name, category, stars, forks, delta_stars, snapshot_date, raw_metadata
		FROM trending_snapshots
		WHERE snapshot_date >= $1
		ORDER BY category, stars DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*model.Snapshot
	for rows.Next() {
		snap, err := scanSnapshotRows(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

// GetStarDelta 获取项目近 N 天的 star 增量
func (s *Store) GetStarDelta(ctx context.Context, projectName string, days int) int {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	var total int
	s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(delta_stars), 0)
		FROM trending_snapshots
		WHERE project_name = $1 AND snapshot_date >= $2
	`, projectName, since).Scan(&total)
	return total
}

// CleanupOldSnapshots 清理超过 N 天的快照
func (s *Store) CleanupOldSnapshots(ctx context.Context, days int) error {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	return s.db.Exec(ctx, `
		DELETE FROM trending_snapshots
		WHERE snapshot_date < $1
		AND project_name NOT IN (
			SELECT DISTINCT project_name
			FROM trending_snapshots
			WHERE snapshot_date >= $1
		)
	`, since)
}

// GetAnalysisCache 获取 LLM 缓存
func (s *Store) GetAnalysisCache(ctx context.Context, readmeMD5 string) (*model.CachedAnalysis, error) {
	row := s.db.QueryRow(ctx, `
		SELECT readme_md5, llm_response, created_at, expires_at
		FROM analysis_cache
		WHERE readme_md5 = $1 AND expires_at > NOW()
	`, readmeMD5)

	var cached model.CachedAnalysis
	var llmResponse []byte
	err := row.Scan(&cached.ReadmeMD5, &llmResponse, &cached.CreatedAt, &cached.ExpiresAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(llmResponse, &cached.LLMResponse); err != nil {
		return nil, fmt.Errorf("unmarshal llm_response: %w", err)
	}
	return &cached, nil
}

// SetAnalysisCache 设置 LLM 缓存
func (s *Store) SetAnalysisCache(ctx context.Context, readmeMD5 string, response *model.LLMResponse, ttl time.Duration) error {
	llmResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal llm_response: %w", err)
	}

	return s.db.Exec(ctx, `
		INSERT INTO analysis_cache (readme_md5, llm_response, expires_at)
		VALUES ($1, $2, NOW() + $3::INTERVAL)
		ON CONFLICT (readme_md5)
		DO UPDATE SET llm_response = $2, expires_at = NOW() + $3::INTERVAL
	`, readmeMD5, llmResponse, fmt.Sprintf("%d seconds", int(ttl.Seconds())))
}

// scanSnapshot 扫描单行 Snapshot
func scanSnapshot(row db.RowScanner) (*model.Snapshot, error) {
	var snap model.Snapshot
	var rawMeta []byte
	err := row.Scan(&snap.ProjectName, &snap.Category, &snap.Stars, &snap.Forks,
		&snap.DeltaStars, &snap.SnapshotDate, &rawMeta)
	if err != nil {
		return nil, err
	}
	if rawMeta != nil {
		json.Unmarshal(rawMeta, &snap.RawMetadata)
	}
	return &snap, nil
}

// scanSnapshotRows 从 Rows 扫描 Snapshot
func scanSnapshotRows(rows db.Rows) (*model.Snapshot, error) {
	var snap model.Snapshot
	var rawMeta []byte
	err := rows.Scan(&snap.ProjectName, &snap.Category, &snap.Stars, &snap.Forks,
		&snap.DeltaStars, &snap.SnapshotDate, &rawMeta)
	if err != nil {
		return nil, err
	}
	if rawMeta != nil {
		json.Unmarshal(rawMeta, &snap.RawMetadata)
	}
	return &snap, nil
}
