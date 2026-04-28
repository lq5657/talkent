package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

type Analyzer interface {
	Analyze(ctx context.Context, sessionID, roleDesc, scenario string, messages []store.Message, dimensions []role.Dimension) (*AnalysisResult, error)
}

type Service struct {
	store        *store.AnalysisStore
	engine       Analyzer
	sessionStore *store.SessionStore
	logger       *slog.Logger
}

func NewService(store *store.AnalysisStore, engine Analyzer, sessionStore *store.SessionStore, logger *slog.Logger) *Service {
	return &Service{
		store:        store,
		engine:       engine,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (s *Service) TriggerAnalysis(ctx context.Context, sessionID string, trigger string) (*AnalysisResult, *store.AnalysisReport, error) {
	sess, err := s.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, nil, ErrSessionNotFound
	}
	if sess.Status != "completed" {
		return nil, nil, ErrSessionNotCompleted
	}

	s.logger.Info("analysis triggered", "session_id", sessionID, "trigger", trigger)

	messages, err := s.sessionStore.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("list messages: %w", err)
	}

	var rc roleConfigData
	if err := json.Unmarshal([]byte(sess.RoleConfig), &rc); err != nil {
		return nil, nil, fmt.Errorf("unmarshal role config: %w", err)
	}

	var dims []role.Dimension
	if err := json.Unmarshal([]byte(sess.Dimensions), &dims); err != nil {
		return nil, nil, fmt.Errorf("unmarshal dimensions: %w", err)
	}

	result, err := s.engine.Analyze(ctx, sessionID, rc.Description, rc.Scenario, messages, dims)
	if err != nil {
		return nil, nil, err
	}

	dimResultsJSON, err := json.Marshal(result.DimensionResults)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal dimension results: %w", err)
	}

	report := &store.AnalysisReport{
		SessionID:        sessionID,
		DimensionResults: string(dimResultsJSON),
		MarkdownContent:  result.Markdown,
		ModelUsed:        result.ModelUsed,
		CreatedAt:        time.Now(),
	}

	if err := s.store.CreateReport(ctx, report); err != nil {
		return nil, nil, fmt.Errorf("save report: %w", err)
	}

	s.logger.Info("analysis report saved", "session_id", sessionID, "report_id", report.ID, "model_used", result.ModelUsed)

	return result, report, nil
}

func (s *Service) GetLatestReport(ctx context.Context, sessionID string) (*store.AnalysisReport, error) {
	report, err := s.store.GetLatestReport(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get latest report: %w", err)
	}
	return report, nil
}

func (s *Service) ListReports(ctx context.Context, sessionID string) ([]store.AnalysisReport, error) {
	reports, err := s.store.ListReports(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	return reports, nil
}

type roleConfigData struct {
	Description string `json:"description"`
	Scenario    string `json:"scenario"`
}
