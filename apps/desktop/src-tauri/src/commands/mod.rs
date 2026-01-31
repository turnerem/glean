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
    pub tags: Vec<String>,
    pub origin: Origin,
}

#[derive(Debug, Deserialize)]
pub struct UpdateNoteRequest {
    pub id: String,
    pub content: String,
    pub tags: Vec<String>,
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
    db.create_note(request.content, request.tags, request.origin)
        .map_err(Into::into)
}

#[tauri::command]
pub fn update_note(request: UpdateNoteRequest) -> Result<(), CommandError> {
    let db = db::get_db();
    db.update_note(&request.id, request.content, request.tags)
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
