use chrono::{DateTime, Utc};
use rusqlite::{params, Connection};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::sync::Mutex;
use tauri::{AppHandle, Manager};
use thiserror::Error;
use uuid::Uuid;

#[derive(Error, Debug)]
pub enum DbError {
    #[error("Database error: {0}")]
    Sqlite(#[from] rusqlite::Error),
    #[error("Serialization error: {0}")]
    Serde(#[from] serde_json::Error),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Note {
    pub id: String,
    pub content: String,
    pub image_data: Option<String>,
    pub tags: Vec<String>,
    pub origin: Origin,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub sync_version: i64,
    pub is_deleted: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Origin {
    #[serde(rename = "type")]
    pub origin_type: String,
    pub url: Option<String>,
    pub title: Option<String>,
    pub book_title: Option<String>,
    pub chapter: Option<String>,
    pub page: Option<String>,
    pub raw_input: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Tag {
    pub id: String,
    pub name: String,
    pub color: Option<String>,
    pub created_at: DateTime<Utc>,
}

pub struct Database {
    conn: Mutex<Connection>,
}

impl std::fmt::Debug for Database {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Database").finish()
    }
}

impl Database {
    pub fn new(path: PathBuf) -> Result<Self, DbError> {
        let conn = Connection::open(path)?;
        Ok(Self {
            conn: Mutex::new(conn),
        })
    }

    pub fn init_schema(&self) -> Result<(), DbError> {
        let conn = self.conn.lock().unwrap();
        conn.execute_batch(
            r#"
            CREATE TABLE IF NOT EXISTS notes (
                id TEXT PRIMARY KEY,
                content TEXT NOT NULL,
                image_data TEXT,
                tags TEXT NOT NULL DEFAULT '[]',
                origin TEXT NOT NULL DEFAULT '{}',
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL,
                sync_version INTEGER NOT NULL DEFAULT 0,
                is_deleted INTEGER NOT NULL DEFAULT 0
            );

            CREATE TABLE IF NOT EXISTS tags (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL UNIQUE,
                color TEXT,
                created_at TEXT NOT NULL
            );

            CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at);
            CREATE INDEX IF NOT EXISTS idx_notes_is_deleted ON notes(is_deleted);
            "#,
        )?;

        // Migration: Add image_data column if it doesn't exist
        let _ = conn.execute("ALTER TABLE notes ADD COLUMN image_data TEXT", []);

        Ok(())
    }

    pub fn get_notes(&self) -> Result<Vec<Note>, DbError> {
        let conn = self.conn.lock().unwrap();
        let mut stmt = conn.prepare(
            "SELECT id, content, image_data, tags, origin, created_at, updated_at, sync_version, is_deleted
             FROM notes WHERE is_deleted = 0 ORDER BY created_at DESC",
        )?;

        let notes = stmt
            .query_map([], |row| {
                let tags_json: String = row.get(3)?;
                let origin_json: String = row.get(4)?;
                Ok(Note {
                    id: row.get(0)?,
                    content: row.get(1)?,
                    image_data: row.get(2)?,
                    tags: serde_json::from_str(&tags_json).unwrap_or_default(),
                    origin: serde_json::from_str(&origin_json).unwrap_or_default(),
                    created_at: row.get(5)?,
                    updated_at: row.get(6)?,
                    sync_version: row.get(7)?,
                    is_deleted: row.get(8)?,
                })
            })?
            .collect::<Result<Vec<_>, _>>()?;

        Ok(notes)
    }

    pub fn create_note(&self, content: String, image_data: Option<String>, tags: Vec<String>, origin: Origin) -> Result<Note, DbError> {
        let conn = self.conn.lock().unwrap();
        let id = Uuid::now_v7().to_string();
        let now = Utc::now();
        let tags_json = serde_json::to_string(&tags)?;
        let origin_json = serde_json::to_string(&origin)?;

        conn.execute(
            "INSERT INTO notes (id, content, image_data, tags, origin, created_at, updated_at, sync_version, is_deleted)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, 0, 0)",
            params![id, content, image_data, tags_json, origin_json, now, now],
        )?;

        Ok(Note {
            id,
            content,
            image_data,
            tags,
            origin,
            created_at: now,
            updated_at: now,
            sync_version: 0,
            is_deleted: false,
        })
    }

    pub fn update_note(&self, id: &str, content: String, tags: Vec<String>, origin: Option<Origin>) -> Result<(), DbError> {
        let conn = self.conn.lock().unwrap();
        let now = Utc::now();
        let tags_json = serde_json::to_string(&tags)?;

        if let Some(origin) = origin {
            let origin_json = serde_json::to_string(&origin)?;
            conn.execute(
                "UPDATE notes SET content = ?1, tags = ?2, origin = ?3, updated_at = ?4, sync_version = sync_version + 1
                 WHERE id = ?5",
                params![content, tags_json, origin_json, now, id],
            )?;
        } else {
            conn.execute(
                "UPDATE notes SET content = ?1, tags = ?2, updated_at = ?3, sync_version = sync_version + 1
                 WHERE id = ?4",
                params![content, tags_json, now, id],
            )?;
        }

        Ok(())
    }

    pub fn delete_note(&self, id: &str) -> Result<(), DbError> {
        let conn = self.conn.lock().unwrap();
        let now = Utc::now();

        conn.execute(
            "UPDATE notes SET is_deleted = 1, updated_at = ?1, sync_version = sync_version + 1 WHERE id = ?2",
            params![now, id],
        )?;

        Ok(())
    }

    pub fn get_tags(&self) -> Result<Vec<Tag>, DbError> {
        let conn = self.conn.lock().unwrap();
        let mut stmt = conn.prepare("SELECT id, name, color, created_at FROM tags ORDER BY name")?;

        let tags = stmt
            .query_map([], |row| {
                Ok(Tag {
                    id: row.get(0)?,
                    name: row.get(1)?,
                    color: row.get(2)?,
                    created_at: row.get(3)?,
                })
            })?
            .collect::<Result<Vec<_>, _>>()?;

        Ok(tags)
    }

    pub fn create_tag(&self, name: String, color: Option<String>) -> Result<Tag, DbError> {
        let conn = self.conn.lock().unwrap();
        let id = Uuid::now_v7().to_string();
        let now = Utc::now();

        conn.execute(
            "INSERT INTO tags (id, name, color, created_at) VALUES (?1, ?2, ?3, ?4)",
            params![id, name, color, now],
        )?;

        Ok(Tag {
            id,
            name,
            color,
            created_at: now,
        })
    }
}

// Global database instance
static DB: std::sync::OnceLock<Database> = std::sync::OnceLock::new();

pub async fn init(app: &AppHandle) -> Result<(), DbError> {
    let app_data_dir = app.path().app_data_dir().expect("Failed to get app data dir");
    std::fs::create_dir_all(&app_data_dir)?;

    let db_path = app_data_dir.join("glean.db");
    log::info!("Initializing database at {:?}", db_path);

    let db = Database::new(db_path)?;
    db.init_schema()?;

    DB.set(db).expect("Database already initialized");
    Ok(())
}

pub fn get_db() -> &'static Database {
    DB.get().expect("Database not initialized")
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_db() -> Database {
        // Use in-memory SQLite for tests
        let conn = Connection::open_in_memory().unwrap();
        let db = Database {
            conn: Mutex::new(conn),
        };
        db.init_schema().unwrap();
        db
    }

    #[test]
    fn test_create_and_get_note() {
        let db = create_test_db();

        let origin = Origin {
            origin_type: "url".to_string(),
            url: Some("https://example.com".to_string()),
            title: Some("Example".to_string()),
            ..Default::default()
        };

        let note = db
            .create_note("Test content".to_string(), None, vec!["tag1".to_string()], origin)
            .unwrap();

        assert_eq!(note.content, "Test content");
        assert_eq!(note.tags, vec!["tag1"]);
        assert_eq!(note.origin.origin_type, "url");
        assert!(!note.is_deleted);

        let notes = db.get_notes().unwrap();
        assert_eq!(notes.len(), 1);
        assert_eq!(notes[0].id, note.id);
    }

    #[test]
    fn test_update_note() {
        let db = create_test_db();

        let origin = Origin::default();
        let note = db
            .create_note("Original".to_string(), None, vec![], origin)
            .unwrap();

        db.update_note(&note.id, "Updated".to_string(), vec!["new_tag".to_string()], None)
            .unwrap();

        let notes = db.get_notes().unwrap();
        assert_eq!(notes[0].content, "Updated");
        assert_eq!(notes[0].tags, vec!["new_tag"]);
    }

    #[test]
    fn test_delete_note() {
        let db = create_test_db();

        let origin = Origin::default();
        let note = db
            .create_note("To delete".to_string(), None, vec![], origin)
            .unwrap();

        db.delete_note(&note.id).unwrap();

        // Deleted notes should not appear in get_notes
        let notes = db.get_notes().unwrap();
        assert_eq!(notes.len(), 0);
    }

    #[test]
    fn test_create_and_get_tags() {
        let db = create_test_db();

        let tag = db.create_tag("work".to_string(), Some("#ff0000".to_string())).unwrap();

        assert_eq!(tag.name, "work");
        assert_eq!(tag.color, Some("#ff0000".to_string()));

        let tags = db.get_tags().unwrap();
        assert_eq!(tags.len(), 1);
        assert_eq!(tags[0].name, "work");
    }

    #[test]
    fn test_multiple_notes_ordering() {
        let db = create_test_db();

        let origin = Origin::default();
        db.create_note("First".to_string(), None, vec![], origin.clone()).unwrap();
        db.create_note("Second".to_string(), None, vec![], origin.clone()).unwrap();
        db.create_note("Third".to_string(), None, vec![], origin).unwrap();

        let notes = db.get_notes().unwrap();
        assert_eq!(notes.len(), 3);
        // Notes should be ordered by created_at DESC (newest first)
        assert_eq!(notes[0].content, "Third");
        assert_eq!(notes[1].content, "Second");
        assert_eq!(notes[2].content, "First");
    }
}
