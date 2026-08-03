use std::io::{BufRead, BufReader, Write};
use std::path::PathBuf;
use std::process::{ChildStdin, Command, Stdio};
use std::sync::{Mutex, MutexGuard, OnceLock};

#[cfg(windows)]
use std::os::windows::process::CommandExt;

use serde::Serialize;
use tauri::ipc::Channel;

#[cfg(windows)]
const CREATE_NO_WINDOW: u32 = 0x0800_0000;

const TAI_BIN: &[u8] = include_bytes!("../binaries/tai-x86_64-pc-windows-msvc.exe");
const TAI_NAME: &str = "tai-cli.exe";

/// The word written to a running CLI's standard input to ask it to stop. This app has no
/// console, so on Windows it can't send the CLI an interrupt signal; the CLI watches its
/// stdin for this word instead, and stops at its next safe point.
const CANCEL_COMMAND: &[u8] = b"cancel\n";

static TAI_PATH: OnceLock<PathBuf> = OnceLock::new();

/// The standard input of the action currently running, when it was started as cancellable.
/// Only one cancellable action runs at a time - the interface won't start another while one
/// is in flight - and the short `check`/`list` probes never register here, so a cancellation
/// can't land on the wrong process.
static RUNNING: Mutex<Option<ChildStdin>> = Mutex::new(None);

/// Access the running action. A lock is only ever held for a couple of statements, so a
/// poisoned one means an unrelated thread panicked; recover from it rather than spreading it.
fn running() -> MutexGuard<'static, Option<ChildStdin>> {
    RUNNING.lock().unwrap_or_else(|error| error.into_inner())
}

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
async fn run_tai(
    args: Vec<String>,
    cancellable: bool,
    on_event: Channel<TaiEvent>,
) -> Result<(), String> {
    let path = TAI_PATH
        .get()
        .ok_or("tai binary not initialized")?
        .clone();

    tauri::async_runtime::spawn_blocking(move || -> Result<(), String> {
        let mut command = Command::new(&path);
        command
            .args(&args)
            // Piped rather than inherited: this is the channel `cancel_tai` writes to.
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        #[cfg(windows)]
        command.creation_flags(CREATE_NO_WINDOW);
        let mut child = command.spawn().map_err(|error| error.to_string())?;

        let stdin = child.stdin.take().ok_or("failed to capture stdin")?;
        let stdout = child.stdout.take().ok_or("failed to capture stdout")?;
        let stderr = child.stderr.take().ok_or("failed to capture stderr")?;

        // Anything else drops its stdin here, closing the pipe. The CLI treats that as
        // "no one is going to ask me to stop", which is exactly right.
        if cancellable {
            *running() = Some(stdin);
        }

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

        if cancellable {
            *running() = None;
        }

        let _ = on_event.send(TaiEvent::Exit(status.code().unwrap_or(-1)));
        Ok(())
    })
    .await
    .map_err(|error| error.to_string())?
}

/// Ask the running action to stop. The CLI winds down at its next safe point, logs why, and
/// exits on its own, so the run's ordinary output and exit event still arrive. Cancelling
/// when nothing is running, or twice over, does nothing.
#[tauri::command]
fn cancel_tai() {
    // Take the handle before writing, so the lock isn't held across the I/O.
    let action = running().take();
    if let Some(mut stdin) = action {
        let _ = stdin.write_all(CANCEL_COMMAND);
        let _ = stdin.flush();
    }
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
        .invoke_handler(tauri::generate_handler![run_tai, cancel_tai])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
