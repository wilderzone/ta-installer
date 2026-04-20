use std::path::PathBuf;

#[tauri::command]
fn get_exe_path() -> Result<PathBuf, String> {
    std::env::current_exe()
        .map_err(|error| format!("{error}"))?
        .parent()
        .map(|path| path.to_path_buf())
        .ok_or("Executable has no parent directory".to_string())
}

#[tauri::command]
fn write_sidecar(data: Vec<u8>) -> Result<(), String> {
    let dest = std::env::current_exe()
        .map_err(|error| error.to_string())?
        .parent()
        .ok_or("Executable has no parent directory")?
        .join("tai-cli.exe");
    std::fs::write(dest, data).map_err(|error| error.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            get_exe_path,
            write_sidecar,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
