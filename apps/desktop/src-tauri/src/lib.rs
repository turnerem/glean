mod commands;
mod db;
mod origin;
mod sync;

use log::info;

pub fn run() {
    env_logger::init();

    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_global_shortcut::Builder::new().build())
        .plugin(tauri_plugin_clipboard_manager::init())
        .setup(|app| {
            info!("Glean starting up");

            // Initialize database
            let app_handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                if let Err(e) = db::init(&app_handle).await {
                    log::error!("Failed to initialize database: {}", e);
                }
            });

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            // Note commands
            commands::get_notes,
            commands::create_note,
            commands::update_note,
            commands::delete_note,
            // Tag commands
            commands::get_tags,
            commands::create_tag,
            // Origin detection
            commands::detect_origin,
            // Sync commands
            commands::sync_get_status,
            commands::sync_configure,
            commands::sync_login,
            commands::sync_register,
            commands::sync_logout,
            commands::sync_now,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
