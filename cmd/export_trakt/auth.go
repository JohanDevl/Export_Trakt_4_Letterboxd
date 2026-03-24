package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// runInteractiveAuth performs interactive OAuth authentication
func runInteractiveAuth(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)

	fmt.Println("🔑 Starting Interactive OAuth Authentication")
	fmt.Println("==========================================")

	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials:")
		fmt.Println("1. Go to https://trakt.tv/oauth/applications")
		fmt.Println("2. Create a new application or modify existing one")
		fmt.Println("3. Set client_id and client_secret in your config file")
		fmt.Printf("4. Set redirect_uri to: %s\n", cfg.Auth.RedirectURI)
		return fmt.Errorf("missing API credentials")
	}

	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)

	// Start local callback server
	callbackURL, codeChan, errChan, err := oauthMgr.StartLocalCallbackServer()
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}

	fmt.Printf("🌐 Local callback server started at: %s\n", callbackURL)

	// Generate authorization URL
	authURL, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}

	fmt.Println("\n📋 NEXT STEPS:")
	fmt.Println("1. Open the following URL in your browser:")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Println("2. Authorize the application on Trakt.tv")
	fmt.Println("3. You will be redirected back automatically")
	fmt.Println("\nWaiting for authorization...")

	// Wait for authorization code or error
	select {
	case code := <-codeChan:
		fmt.Println("✅ Authorization code received!")

		// Exchange code for token
		token, err := oauthMgr.ExchangeCodeForToken(code, state, state)
		if err != nil {
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}

		// Store token
		if err := tokenManager.StoreToken(token); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}

		fmt.Println("🎉 Authentication successful!")
		fmt.Printf("📅 Token expires: %s\n", oauthMgr.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"))
		fmt.Println("🔄 Automatic refresh is enabled")

		return nil

	case err := <-errChan:
		return fmt.Errorf("authentication error: %w", err)

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timeout after 5 minutes")
	}
}

// showTokenStatus displays the current token status
func showTokenStatus(tokenManager *auth.TokenManager) error {
	fmt.Println("🔍 Token Status Check")
	fmt.Println("=====================")

	status, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get token status: %w", err)
	}

	fmt.Println(status.String())

	if status.Error != "" {
		fmt.Printf("\n❌ Error: %s\n", status.Error)
	}

	if status.Message != "" {
		fmt.Printf("\n💡 Info: %s\n", status.Message)
	}

	if !status.HasToken {
		fmt.Println("\n🆘 No token found. Run 'auth' command to authenticate:")
		fmt.Println("   docker exec -it <container> /app/export-trakt auth")
	}

	return nil
}

// refreshToken manually refreshes the access token
func refreshToken(tokenManager *auth.TokenManager, log logger.Logger) error {
	fmt.Println("🔄 Refreshing Access Token")
	fmt.Println("===========================")

	// Check current status first
	status, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get token status: %w", err)
	}

	if !status.HasToken {
		fmt.Println("❌ No token to refresh. Run 'auth' command first.")
		return fmt.Errorf("no token available")
	}

	if !status.HasRefreshToken {
		fmt.Println("❌ No refresh token available. Re-authentication required.")
		fmt.Println("Run: auth")
		return fmt.Errorf("no refresh token available")
	}

	if err := tokenManager.RefreshToken(); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Show new status
	newStatus, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get new token status: %w", err)
	}

	fmt.Println("✅ Token refreshed successfully!")
	fmt.Printf("📅 New expiry: %s\n", newStatus.ExpiresAt.Format("2006-01-02 15:04:05"))

	return nil
}

// clearTokens removes all stored tokens
func clearTokens(tokenManager *auth.TokenManager, log logger.Logger) error {
	fmt.Println("🗑️  Clearing Stored Tokens")
	fmt.Println("===========================")

	if err := tokenManager.ClearToken(); err != nil {
		return fmt.Errorf("failed to clear tokens: %w", err)
	}

	fmt.Println("✅ All tokens cleared successfully!")
	fmt.Println("💡 Run 'auth' command to re-authenticate when needed.")

	return nil
}

// showAuthURL generates and displays the OAuth authentication URL
func showAuthURL(cfg *config.Config, log logger.Logger) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)

	fmt.Println("🔗 OAuth Authentication URL Generator")
	fmt.Println("=====================================")

	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials:")
		fmt.Println("1. Go to https://trakt.tv/oauth/applications")
		fmt.Println("2. Create a new application or modify existing one")
		fmt.Println("3. Set client_id and client_secret in your config file")
		fmt.Printf("4. Set redirect_uri to: %s\n", cfg.Auth.RedirectURI)
		return fmt.Errorf("missing API credentials")
	}

	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)

	// Generate authorization URL
	authURL, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}

	fmt.Println("\n🚀 AUTHENTICATION STEPS:")
	fmt.Println("1. Copy and open this URL in your browser:")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Println("2. Authorize the application on Trakt.tv")
	fmt.Println("3. You will be redirected to localhost - this is normal")
	fmt.Println("4. Copy the 'code' parameter from the URL")
	fmt.Println("5. Run the interactive auth command:")
	fmt.Printf("   docker run --rm -v \"$(pwd)/config:/app/config\" -p %d:%d trakt-exporter auth\n", cfg.Auth.CallbackPort, cfg.Auth.CallbackPort)
	fmt.Println("\n💾 State (for security):", state)
	fmt.Println("\n💡 This URL is valid for 10 minutes.")

	return nil
}

// authenticateWithCode performs OAuth authentication using a provided authorization code
func authenticateWithCode(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, authCode string) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)

	fmt.Println("🔑 Manual OAuth Authentication with Code")
	fmt.Println("=========================================")

	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials in config.toml")
		return fmt.Errorf("missing API credentials")
	}

	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	fmt.Printf("🔐 Authorization Code: %s\n", authCode)

	// Generate a state for this manual authentication (not validated since we're not using callback)
	_, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	fmt.Println("\n🔄 Exchanging authorization code for tokens...")

	// Exchange code for token (we'll use the generated state)
	token, err := oauthMgr.ExchangeCodeForToken(authCode, state, state)
	if err != nil {
		fmt.Printf("❌ Token exchange failed: %s\n", err.Error())
		fmt.Println("\n💡 Possible reasons:")
		fmt.Println("   - Authorization code has expired (they expire quickly)")
		fmt.Println("   - Authorization code has already been used")
		fmt.Println("   - Redirect URI mismatch in Trakt.tv app settings")
		fmt.Printf("   - Expected redirect URI: %s\n", cfg.Auth.RedirectURI)
		return err
	}

	// Store the token
	if err := tokenManager.StoreToken(token); err != nil {
		fmt.Printf("❌ Failed to store token: %s\n", err.Error())
		return err
	}

	fmt.Println("✅ Authentication successful!")
	fmt.Printf("📅 Token expires: %s\n", oauthMgr.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"))
	fmt.Println("🔄 Automatic refresh is enabled")
	fmt.Println("\n💡 You can now run export commands normally.")

	return nil
}

// fixCredentialsPermissions fixes file permissions for credentials storage
func fixCredentialsPermissions(cfg *config.Config, log logger.Logger) error {
	credentialsPath := "./config/credentials.enc"

	fmt.Printf("🔧 Fixing credentials file permissions...\n\n")

	// Check if file exists
	info, err := os.Stat(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("✅ Credentials file doesn't exist yet - no action needed.\n")
			fmt.Printf("   File will be created with proper permissions when you authenticate.\n")
			return nil
		}
		return fmt.Errorf("failed to check credentials file: %w", err)
	}

	currentMode := info.Mode()
	fmt.Printf("📋 Current file permissions: %o\n", currentMode&os.ModePerm)

	// Check Docker environment
	isDocker := false
	if _, err := os.Stat("/.dockerenv"); err == nil {
		isDocker = true
		fmt.Printf("🐳 Detected Docker environment\n")
	}

	// Determine target permissions
	targetMode := os.FileMode(0600)
	if isDocker {
		// In Docker, we might need more relaxed permissions
		if currentMode&0077 != 0 && currentMode&0044 == 0 {
			fmt.Printf("✅ File permissions are acceptable for Docker environment.\n")
			return nil
		}
		// Try to set more restrictive permissions, but accept failure in Docker
		targetMode = os.FileMode(0644)
	}

	fmt.Printf("🎯 Target permissions: %o\n", targetMode)

	// Try to change permissions
	if err := os.Chmod(credentialsPath, targetMode); err != nil {
		if isDocker {
			fmt.Printf("⚠️  Warning: Could not change file permissions in Docker environment.\n")
			fmt.Printf("   This is normal - Docker handles file permissions differently.\n")
			fmt.Printf("   Your credentials should still work properly.\n")
			return nil
		}
		return fmt.Errorf("failed to change file permissions: %w", err)
	}

	// Verify the change
	newInfo, err := os.Stat(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to verify permissions change: %w", err)
	}

	newMode := newInfo.Mode()
	fmt.Printf("✅ Permissions updated successfully: %o\n", newMode&os.ModePerm)

	fmt.Printf("\n💡 Tips:\n")
	fmt.Printf("   - If you're still having issues, try using the 'env' keyring backend\n")
	fmt.Printf("   - Set TRAKT_CLIENT_ID and TRAKT_CLIENT_SECRET environment variables\n")
	fmt.Printf("   - Update config.toml: keyring_backend = \"env\"\n")

	return nil
}
