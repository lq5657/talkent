package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AnalysisReport struct {
	ID               int64
	SessionID        string
	DimensionResults string
	MarkdownContent  string
	ModelUsed        string
	CreatedAt        time.Time
}

type AnalysisStore struct {
	db *sql.DB
}

func NewAnalysisStore(db *sql.DB) *AnalysisStore {
	return &AnalysisStore{db: db}
}

func (s *AnalysisStore) CreateReport(ctx context.Context, report *AnalysisReport) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO analysis_reports (session_id, dimension_results, markdown_content, model_used, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		report.SessionID, report.DimensionResults, report.MarkdownContent, report.ModelUsed, report.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert analysis report: %w", err)
	}
	report.ID, _ = result.LastInsertId()
	return nil
}

func (s *AnalysisStore) GetLatestReport(ctx context.Context, sessionID string) (*AnalysisReport, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, session_id, dimension_results, markdown_content, model_used, created_at
		 FROM analysis_reports WHERE session_id = ? ORDER BY created_at DESC LIMIT 1`, sessionID,
	)
	var r AnalysisReport
	if err := row.Scan(&r.ID, &r.SessionID, &r.DimensionResults, &r.MarkdownContent, &r.ModelUsed, &r.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest report: %w", err)
	}
	return &r, nil
}

func (s *AnalysisStore) ListReports(ctx context.Context, sessionID string) ([]AnalysisReport, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, dimension_results, markdown_content, model_used, created_at
		 FROM analysis_reports WHERE session_id = ? ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []AnalysisReport
	for rows.Next() {
		var r AnalysisReport
		if err := rows.Scan(&r.ID, &r.SessionID, &r.DimensionResults, &r.MarkdownContent, &r.ModelUsed, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}
