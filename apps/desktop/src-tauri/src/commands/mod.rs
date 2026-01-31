use crate::db::{self, Note, Origin, Tag};
use crate::origin::detect;
use crate::sync::{self, SyncResponse};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct CommandError {
    message: String,
}

impl From<db::DbError> for CommandError {
    fn from(err: db::DbError) -> Self {
        CommandError {
            message: err.to_string(),
        }
    }
}

impl From<crate::origin::OriginError> for CommandError {
    fn from(err: crate::origin::OriginError) -> Self {
        CommandError {
            message: err.to_string(),
        }
    }
}

impl From<sync::SyncError> for CommandError {
    fn from(err: sync::SyncError) -> Self {
        CommandError {
            message: err.to_string(),
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct CreateNoteRequest {
    pub content: String,
    pub image_data: Option<String>,
    pub tags: Vec<String>,
    pub origin: Origin,
}

#[derive(Debug, Deserialize)]
pub struct UpdateNoteRequest {
    pub id: String,
    pub content: String,
    pub tags: Vec<String>,
    pub origin: Option<Origin>,
}

#[derive(Debug, Deserialize)]
pub struct CreateTagRequest {
    pub name: String,
    pub color: Option<String>,
}

#[tauri::command]
pub fn get_notes() -> Result<Vec<Note>, CommandError> {
    let db = db::get_db();
    db.get_notes().map_err(Into::into)
}

#[tauri::command]
pub fn create_note(request: CreateNoteRequest) -> Result<Note, CommandError> {
    let db = db::get_db();
    db.create_note(request.content, request.image_data, request.tags, request.origin)
        .map_err(Into::into)
}

#[tauri::command]
pub fn update_note(request: UpdateNoteRequest) -> Result<(), CommandError> {
    let db = db::get_db();
    db.update_note(&request.id, request.content, request.tags, request.origin)
        .map_err(Into::into)
}

#[tauri::command]
pub fn delete_note(id: String) -> Result<(), CommandError> {
    let db = db::get_db();
    db.delete_note(&id).map_err(Into::into)
}

#[tauri::command]
pub fn get_tags() -> Result<Vec<Tag>, CommandError> {
    let db = db::get_db();
    db.get_tags().map_err(Into::into)
}

#[tauri::command]
pub fn create_tag(request: CreateTagRequest) -> Result<Tag, CommandError> {
    let db = db::get_db();
    db.create_tag(request.name, request.color).map_err(Into::into)
}

#[tauri::command]
pub async fn detect_origin() -> Result<Origin, CommandError> {
    detect().await.map_err(Into::into)
}

/// Simulates Cmd+C keystroke to copy selected text to clipboard.
/// Requires Accessibility permissions on macOS.
#[tauri::command]
pub async fn simulate_copy() -> Result<(), CommandError> {
    #[cfg(target_os = "macos")]
    {
        use std::process::Command;

        // Use AppleScript to simulate Cmd+C
        let output = Command::new("osascript")
            .arg("-e")
            .arg(r#"tell application "System Events" to keystroke "c" using command down"#)
            .output()
            .map_err(|e| CommandError {
                message: format!("Failed to simulate copy: {}", e),
            })?;

        if !output.status.success() {
            return Err(CommandError {
                message: format!(
                    "AppleScript failed: {}",
                    String::from_utf8_lossy(&output.stderr)
                ),
            });
        }

        // Brief delay for clipboard to populate
        tokio::time::sleep(std::time::Duration::from_millis(100)).await;
    }

    #[cfg(not(target_os = "macos"))]
    {
        // On other platforms, do nothing - user must copy manually
        log::warn!("simulate_copy is only supported on macOS");
    }

    Ok(())
}

// Sync commands

#[derive(Debug, Serialize)]
pub struct SyncStatus {
    pub enabled: bool,
    pub authenticated: bool,
}

#[tauri::command]
pub async fn sync_get_status() -> SyncStatus {
    let client = sync::get_client();
    SyncStatus {
        enabled: client.is_enabled().await,
        authenticated: client.is_authenticated().await,
    }
}

#[tauri::command]
pub async fn sync_configure(server_url: String, enabled: bool) -> Result<(), CommandError> {
    let client = sync::get_client();
    client.configure(server_url, enabled).await;
    Ok(())
}

#[tauri::command]
pub async fn sync_login(email: String, password: String) -> Result<(), CommandError> {
    let client = sync::get_client();
    client.login(&email, &password).await.map_err(Into::into)
}

#[tauri::command]
pub async fn sync_register(email: String, password: String) -> Result<(), CommandError> {
    let client = sync::get_client();
    client.register(&email, &password).await.map_err(Into::into)
}

#[tauri::command]
pub async fn sync_logout() -> Result<(), CommandError> {
    let client = sync::get_client();
    client.logout().await;
    Ok(())
}

#[tauri::command]
pub async fn sync_now() -> Result<SyncResponse, CommandError> {
    let db = db::get_db();
    let client = sync::get_client();

    // Get all local notes and tags
    let notes = db.get_notes()?;
    let tags = db.get_tags()?;

    // Sync with server
    let response = client.sync(notes, tags).await?;

    // TODO: Merge server data back to local DB
    // For now, just return the response

    Ok(response)
}
