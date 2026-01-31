//! Sync module - handles synchronization with the Go backend.
//!
//! Strategy: Server-authoritative with optimistic local updates
//! - On save: write locally, queue for sync
//! - Sync: POST to server, server returns canonical version
//! - Conflicts: Server wins, preserve local copy for review
//! - Offline: Queue operations, replay on reconnect

use crate::db::{Note, Tag};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum SyncError {
    #[error("Network error: {0}")]
    Network(#[from] reqwest::Error),
    #[error("Authentication required")]
    AuthRequired,
    #[error("Invalid credentials")]
    InvalidCredentials,
    #[error("Server error: {0}")]
    Server(String),
    #[error("Not configured")]
    NotConfigured,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncConfig {
    pub server_url: String,
    pub enabled: bool,
}

impl Default for SyncConfig {
    fn default() -> Self {
        Self {
            server_url: "http://localhost:8080".to_string(),
            enabled: false,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthTokens {
    pub access_token: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AuthRequest {
    email: String,
    password: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AuthResponse {
    token: String,
    user: serde_json::Value,
}

#[derive(Debug, Serialize, Deserialize)]
struct SyncRequest {
    last_sync_version: i64,
    notes: Vec<Note>,
    tags: Vec<Tag>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SyncResponse {
    pub server_version: i64,
    pub notes: Vec<Note>,
    pub tags: Vec<Tag>,
    pub conflicts: Vec<Note>,
}

pub struct SyncClient {
    client: Client,
    config: Arc<RwLock<SyncConfig>>,
    tokens: Arc<RwLock<Option<AuthTokens>>>,
    last_sync_version: Arc<RwLock<i64>>,
}

impl SyncClient {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
            config: Arc::new(RwLock::new(SyncConfig::default())),
            tokens: Arc::new(RwLock::new(None)),
            last_sync_version: Arc::new(RwLock::new(0)),
        }
    }

    pub async fn configure(&self, server_url: String, enabled: bool) {
        let mut config = self.config.write().await;
        config.server_url = server_url;
        config.enabled = enabled;
    }

    pub async fn is_enabled(&self) -> bool {
        self.config.read().await.enabled
    }

    pub async fn is_authenticated(&self) -> bool {
        self.tokens.read().await.is_some()
    }

    pub async fn login(&self, email: &str, password: &str) -> Result<(), SyncError> {
        let config = self.config.read().await;
        if !config.enabled {
            return Err(SyncError::NotConfigured);
        }

        let url = format!("{}/api/v1/auth/login", config.server_url);
        let response = self
            .client
            .post(&url)
            .json(&AuthRequest {
                email: email.to_string(),
                password: password.to_string(),
            })
            .send()
            .await?;

        if response.status() == 401 {
            return Err(SyncError::InvalidCredentials);
        }

        if !response.status().is_success() {
            let text = response.text().await.unwrap_or_default();
            return Err(SyncError::Server(text));
        }

        let auth: AuthResponse = response.json().await?;
        let mut tokens = self.tokens.write().await;
        *tokens = Some(AuthTokens {
            access_token: auth.token,
        });

        Ok(())
    }

    pub async fn register(&self, email: &str, password: &str) -> Result<(), SyncError> {
        let config = self.config.read().await;
        if !config.enabled {
            return Err(SyncError::NotConfigured);
        }

        let url = format!("{}/api/v1/auth/register", config.server_url);
        let response = self
            .client
            .post(&url)
            .json(&AuthRequest {
                email: email.to_string(),
                password: password.to_string(),
            })
            .send()
            .await?;

        if response.status() == 409 {
            return Err(SyncError::Server("Email already registered".to_string()));
        }

        if !response.status().is_success() {
            let text = response.text().await.unwrap_or_default();
            return Err(SyncError::Server(text));
        }

        let auth: AuthResponse = response.json().await?;
        let mut tokens = self.tokens.write().await;
        *tokens = Some(AuthTokens {
            access_token: auth.token,
        });

        Ok(())
    }

    pub async fn logout(&self) {
        let mut tokens = self.tokens.write().await;
        *tokens = None;
    }

    pub async fn sync(&self, notes: Vec<Note>, tags: Vec<Tag>) -> Result<SyncResponse, SyncError> {
        let config = self.config.read().await;
        if !config.enabled {
            return Err(SyncError::NotConfigured);
        }

        let tokens = self.tokens.read().await;
        let tokens = tokens.as_ref().ok_or(SyncError::AuthRequired)?;

        let last_version = *self.last_sync_version.read().await;

        let url = format!("{}/api/v1/sync", config.server_url);
        let response = self
            .client
            .post(&url)
            .header("Authorization", format!("Bearer {}", tokens.access_token))
            .json(&SyncRequest {
                last_sync_version: last_version,
                notes,
                tags,
            })
            .send()
            .await?;

        if response.status() == 401 {
            return Err(SyncError::AuthRequired);
        }

        if !response.status().is_success() {
            let text = response.text().await.unwrap_or_default();
            return Err(SyncError::Server(text));
        }

        let sync_response: SyncResponse = response.json().await?;

        // Update last sync version
        let mut version = self.last_sync_version.write().await;
        *version = sync_response.server_version;

        Ok(sync_response)
    }
}

// Global sync client instance
static SYNC_CLIENT: std::sync::OnceLock<SyncClient> = std::sync::OnceLock::new();

pub fn get_client() -> &'static SyncClient {
    SYNC_CLIENT.get_or_init(SyncClient::new)
}
