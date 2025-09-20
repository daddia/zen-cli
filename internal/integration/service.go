package integration

import (
	"context"
	"fmt"
	"sync"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
)

// Service implements the main integration service
type Service struct {
	config    *config.IntegrationsConfig
	logger    logging.Logger
	auth      auth.Manager
	cache     cache.Manager[*TaskSyncRecord]
	providers map[string]IntegrationProvider
	mu        sync.RWMutex
}

// NewService creates a new integration service
func NewService(
	cfg *config.IntegrationsConfig,
	logger logging.Logger,
	authManager auth.Manager,
	cacheManager cache.Manager[*TaskSyncRecord],
) *Service {
	return &Service{
		config:    cfg,
		logger:    logger,
		auth:      authManager,
		cache:     cacheManager,
		providers: make(map[string]IntegrationProvider),
	}
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (IntegrationProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("integration provider '%s' not found", name)
	}

	return provider, nil
}

// RegisterProvider registers a new integration provider
func (s *Service) RegisterProvider(provider IntegrationProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.providers[name] = provider
	s.logger.Info("registered integration provider", "provider", name)

	return nil
}

// ListProviders returns all registered providers
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}

	return providers
}

// IsConfigured checks if integration is properly configured
func (s *Service) IsConfigured() bool {
	if s.config == nil {
		return false
	}

	// Check if a task system is configured
	if s.config.TaskSystem == "" || s.config.TaskSystem == "none" {
		return false
	}

	// Check if the configured provider exists
	s.mu.RLock()
	_, exists := s.providers[s.config.TaskSystem]
	s.mu.RUnlock()

	return exists
}

// GetTaskSystem returns the configured task system of record
func (s *Service) GetTaskSystem() string {
	if s.config == nil {
		return ""
	}
	return s.config.TaskSystem
}

// IsSyncEnabled returns true if sync is enabled
func (s *Service) IsSyncEnabled() bool {
	if s.config == nil {
		return false
	}
	return s.config.SyncEnabled
}

// SyncTask synchronizes a single task
func (s *Service) SyncTask(ctx context.Context, taskID string, opts SyncOptions) (*SyncResult, error) {
	s.logger.Debug("starting task sync", "task_id", taskID, "direction", opts.Direction)

	// Check if integration is configured
	if !s.IsConfigured() {
		return nil, fmt.Errorf("integration not configured")
	}

	// Get the provider
	provider, err := s.GetProvider(s.config.TaskSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Get sync record
	syncRecord, err := s.GetSyncRecord(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("sync record not found for task %s: %w", taskID, err)
	}

	result := &SyncResult{
		TaskID:        taskID,
		Direction:     opts.Direction,
		ChangedFields: []string{},
		Conflicts:     []string{},
		Metadata:      make(map[string]interface{}),
	}

	// Perform sync based on direction
	switch opts.Direction {
	case SyncDirectionPull:
		err = s.pullFromExternal(ctx, provider, syncRecord, result, opts)
	case SyncDirectionPush:
		err = s.pushToExternal(ctx, provider, syncRecord, result, opts)
	case SyncDirectionBidirectional:
		err = s.bidirectionalSync(ctx, provider, syncRecord, result, opts)
	default:
		err = fmt.Errorf("unsupported sync direction: %s", opts.Direction)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		s.logger.Error("task sync failed", "task_id", taskID, "error", err)
		return result, err
	}

	result.Success = true
	result.ExternalID = syncRecord.ExternalID

	// Update sync record
	if err := s.UpdateSyncRecord(ctx, syncRecord); err != nil {
		s.logger.Warn("failed to update sync record", "task_id", taskID, "error", err)
	}

	s.logger.Info("task sync completed", "task_id", taskID, "external_id", syncRecord.ExternalID, "changed_fields", len(result.ChangedFields))

	return result, nil
}

// SyncAllTasks synchronizes all tasks configured for sync
func (s *Service) SyncAllTasks(ctx context.Context, opts SyncOptions) ([]*SyncResult, error) {
	s.logger.Debug("starting sync all tasks", "direction", opts.Direction)

	syncRecords, err := s.ListSyncRecords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync records: %w", err)
	}

	results := make([]*SyncResult, 0, len(syncRecords))

	for _, record := range syncRecords {
		result, err := s.SyncTask(ctx, record.TaskID, opts)
		if err != nil {
			s.logger.Error("failed to sync task", "task_id", record.TaskID, "error", err)
			// Continue with other tasks even if one fails
		}
		if result != nil {
			results = append(results, result)
		}
	}

	s.logger.Info("sync all tasks completed", "total", len(syncRecords), "results", len(results))

	return results, nil
}

// GetSyncRecord retrieves the sync record for a task
func (s *Service) GetSyncRecord(ctx context.Context, taskID string) (*TaskSyncRecord, error) {
	entry, err := s.cache.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("sync record not found for task %s: %w", taskID, err)
	}

	return entry.Data, nil
}

// CreateSyncRecord creates a new sync record
func (s *Service) CreateSyncRecord(ctx context.Context, record *TaskSyncRecord) error {
	if record.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Check if record already exists
	if _, err := s.GetSyncRecord(ctx, record.TaskID); err == nil {
		return fmt.Errorf("sync record already exists for task %s", record.TaskID)
	}

	opts := cache.PutOptions{}
	if err := s.cache.Put(ctx, record.TaskID, record, opts); err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	s.logger.Debug("created sync record", "task_id", record.TaskID, "external_id", record.ExternalID)

	return nil
}

// UpdateSyncRecord updates an existing sync record
func (s *Service) UpdateSyncRecord(ctx context.Context, record *TaskSyncRecord) error {
	if record.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	opts := cache.PutOptions{}
	if err := s.cache.Put(ctx, record.TaskID, record, opts); err != nil {
		return fmt.Errorf("failed to update sync record: %w", err)
	}

	s.logger.Debug("updated sync record", "task_id", record.TaskID, "external_id", record.ExternalID)

	return nil
}

// DeleteSyncRecord deletes a sync record
func (s *Service) DeleteSyncRecord(ctx context.Context, taskID string) error {
	if err := s.cache.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete sync record: %w", err)
	}

	s.logger.Debug("deleted sync record", "task_id", taskID)

	return nil
}

// ListSyncRecords lists all sync records
func (s *Service) ListSyncRecords(ctx context.Context) ([]*TaskSyncRecord, error) {
	// Note: This is a simplified implementation
	// In a production system, you might want to use a different storage mechanism
	// that supports listing all records more efficiently

	// For now, we'll return an empty list as the cache interface doesn't support listing
	// This would need to be enhanced with a proper storage backend for sync records
	s.logger.Debug("listing sync records - simplified implementation")

	return []*TaskSyncRecord{}, nil
}

// pullFromExternal pulls data from external system to Zen
func (s *Service) pullFromExternal(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	if record.ExternalID == "" {
		return fmt.Errorf("external ID not set for task %s", record.TaskID)
	}

	// Get external data
	externalData, err := provider.GetTaskData(ctx, record.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to get external task data: %w", err)
	}

	// Convert to Zen format
	zenData, err := provider.MapToZen(externalData)
	if err != nil {
		return fmt.Errorf("failed to map external data to Zen format: %w", err)
	}

	// TODO: Update Zen task with external data
	// This would integrate with the task management system

	s.logger.Debug("pulled data from external system",
		"task_id", record.TaskID,
		"external_id", record.ExternalID,
		"title", zenData.Title)

	return nil
}

// pushToExternal pushes data from Zen to external system
func (s *Service) pushToExternal(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	// TODO: Get Zen task data
	// This would integrate with the task management system

	// For now, create a placeholder Zen task data
	zenData := &ZenTaskData{
		ID:          record.TaskID,
		Title:       "Sample Task",
		Description: "Sample task description",
		Status:      "In Progress",
		Priority:    "Medium",
	}

	// Convert to external format
	_, err := provider.MapToExternal(zenData)
	if err != nil {
		return fmt.Errorf("failed to map Zen data to external format: %w", err)
	}

	// Update or create external task
	var updatedExternal *ExternalTaskData
	if record.ExternalID != "" {
		updatedExternal, err = provider.UpdateTask(ctx, record.ExternalID, zenData)
	} else {
		updatedExternal, err = provider.CreateTask(ctx, zenData)
		if err == nil {
			record.ExternalID = updatedExternal.ID
		}
	}

	if err != nil {
		return fmt.Errorf("failed to update external task: %w", err)
	}

	s.logger.Debug("pushed data to external system",
		"task_id", record.TaskID,
		"external_id", updatedExternal.ID)

	return nil
}

// bidirectionalSync performs bidirectional synchronization
func (s *Service) bidirectionalSync(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	// For bidirectional sync, we need to compare timestamps and handle conflicts
	// This is a simplified implementation - in production, you'd want more sophisticated conflict resolution

	// First, try to pull from external
	pullOpts := opts
	pullOpts.Direction = SyncDirectionPull
	if err := s.pullFromExternal(ctx, provider, record, result, pullOpts); err != nil {
		s.logger.Warn("failed to pull during bidirectional sync", "error", err)
	}

	// Then, try to push to external
	pushOpts := opts
	pushOpts.Direction = SyncDirectionPush
	if err := s.pushToExternal(ctx, provider, record, result, pushOpts); err != nil {
		s.logger.Warn("failed to push during bidirectional sync", "error", err)
	}

	return nil
}
