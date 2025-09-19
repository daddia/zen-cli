package assets

import (
	"context"

	"github.com/daddia/zen/pkg/auth"
)

// AuthProviderAdapter adapts the shared auth.Manager to the assets AuthProvider interface
type AuthProviderAdapter struct {
	authManager auth.Manager
}

// NewAuthProviderAdapter creates a new adapter for the shared auth manager
func NewAuthProviderAdapter(authManager auth.Manager) AuthProvider {
	return &AuthProviderAdapter{
		authManager: authManager,
	}
}

// Authenticate authenticates with the Git provider
func (a *AuthProviderAdapter) Authenticate(ctx context.Context, provider string) error {
	err := a.authManager.Authenticate(ctx, provider)
	if err != nil {
		// Convert auth errors to asset errors if needed
		if authErr, ok := err.(*auth.Error); ok {
			return &AssetClientError{
				Code:    convertAuthErrorCode(authErr.Code),
				Message: authErr.Message,
				Details: authErr.Details,
			}
		}
		return err
	}
	return nil
}

// GetCredentials returns credentials for the specified provider
func (a *AuthProviderAdapter) GetCredentials(provider string) (string, error) {
	token, err := a.authManager.GetCredentials(provider)
	if err != nil {
		// Convert auth errors to asset errors if needed
		if authErr, ok := err.(*auth.Error); ok {
			return "", &AssetClientError{
				Code:    convertAuthErrorCode(authErr.Code),
				Message: authErr.Message,
				Details: authErr.Details,
			}
		}
		return "", err
	}
	return token, nil
}

// ValidateCredentials validates stored credentials
func (a *AuthProviderAdapter) ValidateCredentials(ctx context.Context, provider string) error {
	err := a.authManager.ValidateCredentials(ctx, provider)
	if err != nil {
		// Convert auth errors to asset errors if needed
		if authErr, ok := err.(*auth.Error); ok {
			return &AssetClientError{
				Code:    convertAuthErrorCode(authErr.Code),
				Message: authErr.Message,
				Details: authErr.Details,
			}
		}
		return err
	}
	return nil
}

// RefreshCredentials refreshes expired credentials if possible
func (a *AuthProviderAdapter) RefreshCredentials(ctx context.Context, provider string) error {
	err := a.authManager.RefreshCredentials(ctx, provider)
	if err != nil {
		// Convert auth errors to asset errors if needed
		if authErr, ok := err.(*auth.Error); ok {
			return &AssetClientError{
				Code:    convertAuthErrorCode(authErr.Code),
				Message: authErr.Message,
				Details: authErr.Details,
			}
		}
		return err
	}
	return nil
}

// convertAuthErrorCode converts auth error codes to asset error codes
func convertAuthErrorCode(authCode auth.ErrorCode) AssetErrorCode {
	switch authCode {
	case auth.ErrorCodeAuthenticationFailed, auth.ErrorCodeInvalidCredentials, auth.ErrorCodeCredentialNotFound:
		return ErrorCodeAuthenticationFailed
	case auth.ErrorCodeNetworkError:
		return ErrorCodeNetworkError
	case auth.ErrorCodeRateLimited:
		return ErrorCodeRateLimited
	case auth.ErrorCodeConfigurationError, auth.ErrorCodeProviderNotSupported:
		return ErrorCodeConfigurationError
	default:
		return ErrorCodeAuthenticationFailed
	}
}
