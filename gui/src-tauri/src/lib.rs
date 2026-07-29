use std::io::{BufRead, BufReader};
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::OnceLock;

#[cfg(windows)]
use std::os::windows::process::CommandExt;

use serde::Serialize;
use tauri::ipc::Channel;

#[cfg(windows)]
const CREATE_NO_WINDOW: u32 = 0x0800_0000;

const TAI_BIN: &[u8] = include_bytes!("../binaries/tai-x86_64-pc-windows-msvc.exe");
const TAI_NAME: &str = "tai-cli.exe";

static TAI_PATH: OnceLock<PathBuf> = OnceLock::new();

fn extract_tai() -> Result<PathBuf, Box<dyn std::error::Error>> {
    let path = std::env::temp_dir().join(TAI_NAME);
    std::fs::write(&path, TAI_BIN)?;
    Ok(path)
}

#[derive(Serialize, Clone)]
#[serde(rename_all = "camelCase", tag = "type", content = "data")]
enum TaiEvent {
    Stdout(String),
    Stderr(String),
    Exit(i32),
}

#[tauri::command]
async fn run_tai(args: Vec<String>, on_event: Channel<TaiEvent>) -> Result<(), String> {
    let path = TAI_PATH
        .get()
        .ok_or("tai binary not initialized")?
        .clone();

    tauri::async_runtime::spawn_blocking(move || -> Result<(), String> {
        let mut command = Command::new(&path);
        command
            .args(&args)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        #[cfg(windows)]
        command.creation_flags(CREATE_NO_WINDOW);
        let mut child = command.spawn().map_err(|error| error.to_string())?;

        let stdout = child.stdout.take().ok_or("failed to capture stdout")?;
        let stderr = child.stderr.take().ok_or("failed to capture stderr")?;

        let stdout_channel = on_event.clone();
        let stdout_thread = std::thread::spawn(move || {
            for line in BufReader::new(stdout).lines().map_while(Result::ok) {
                let _ = stdout_channel.send(TaiEvent::Stdout(line));
            }
        });

        let stderr_channel = on_event.clone();
        let stderr_thread = std::thread::spawn(move || {
            for line in BufReader::new(stderr).lines().map_while(Result::ok) {
                let _ = stderr_channel.send(TaiEvent::Stderr(line));
            }
        });

        let status = child.wait().map_err(|error| error.to_string())?;
        let _ = stdout_thread.join();
        let _ = stderr_thread.join();

        let _ = on_event.send(TaiEvent::Exit(status.code().unwrap_or(-1)));
        Ok(())
    })
    .await
    .map_err(|error| error.to_string())?
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_opener::init())
        .setup(|_app| {
            let path = extract_tai()?;
            let _ = TAI_PATH.set(path);
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![run_tai])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
