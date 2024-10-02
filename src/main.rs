use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::ffi::CString;
use std::ptr::null_mut;
use std::thread::sleep;
use std::time::Duration;
use winapi::shared::minwindef::{BOOL, LPARAM, WPARAM};
use winapi::shared::windef::HWND;
use winapi::um::processthreadsapi::{CreateProcessA, PROCESS_INFORMATION, STARTUPINFOA};
use winapi::um::winuser::*;

#[derive(Serialize, Deserialize)]
struct WizardInfo {
    username: String,
    password: String,
    x_pos: i32,
    y_pos: i32,
}

#[derive(Serialize, Deserialize)]
struct Config {
    file_path: String,
    accounts_data: Vec<WizardInfo>,
}

struct Application {
    config: Config,
}

impl Application {
    fn new(config: Config) -> Self {
        Application { config }
    }

    fn move_window(handle: HWND, x: i32, y: i32) {
        unsafe {
            SetWindowPos(handle, null_mut(), x, y, 0, 0, 0x0001);
        }
    }

    fn get_all_wizard_handles() -> HashMap<HWND, ()> {
        let mut handles = HashMap::new();

        unsafe extern "system" fn enum_windows_callback(hwnd: HWND, lparam: LPARAM) -> BOOL {
            let mut class_name_buf = [0u16; 256];
            GetClassNameW(hwnd, class_name_buf.as_mut_ptr(), 256);
            let class_name = String::from_utf16_lossy(&class_name_buf);

            if class_name.contains("Wizard Graphical Client") {
                let handles = &mut *(lparam as *mut HashMap<HWND, ()>);
                handles.insert(hwnd, ());
            }

            1
        }

        unsafe {
            EnumWindows(
                Some(enum_windows_callback),
                &mut handles as *mut _ as LPARAM,
            );
        }

        handles
    }

    fn open_wizard(&self, filepath: &str) -> Result<PROCESS_INFORMATION, String> {
        let mut si: STARTUPINFOA = unsafe { std::mem::zeroed() };
        si.cb = std::mem::size_of::<STARTUPINFOA>() as u32;
        let mut pi: PROCESS_INFORMATION = unsafe { std::mem::zeroed() };

        let command = format!(
            "cmd /C cd {} && start WizardGraphicalClient.exe -L login.us.wizard101.com 12000",
            filepath
        );
        let c_command = CString::new(command).unwrap();

        let success = unsafe {
            CreateProcessA(
                null_mut(),
                c_command.into_raw(),
                null_mut(),
                null_mut(),
                false as i32,
                SW_HIDE.try_into().unwrap(),
                null_mut(),
                null_mut(),
                &mut si,
                &mut pi,
            )
        };

        if success == 0 {
            Err("Failed to open Wizard101 process".to_string())
        } else {
            Ok(pi)
        }
    }

    fn send_chars(window_handle: HWND, chars: &str) {
        for char in chars.chars() {
            unsafe {
                PostMessageW(window_handle, 0x102, char as WPARAM, 0);
            }
        }
    }

    fn wizard_login(&self, window_handle: HWND, username: &str, password: &str) {
        Application::send_chars(window_handle, username);
        Application::send_chars(window_handle, "\t"); // tab
        Application::send_chars(window_handle, password);
        Application::send_chars(window_handle, "\r"); // enter

        // Set title
        let title = format!("[{}] Wizard101", username);
        let c_title = CString::new(title).unwrap();
        unsafe {
            SetWindowTextW(window_handle, c_title.as_ptr() as *const u16);
        }
    }

    fn run(&self) {
        let target = self.config.accounts_data.len();
        let initial_handles = Self::get_all_wizard_handles();
        let initial_handles_len = initial_handles.len();

        let processes: Vec<_> = (0..target)
            .map(|_| self.open_wizard(&self.config.file_path))
            .collect();

        processes.iter().for_each(|result| match result {
            Ok(_process) => println!("Successfully opened Wizard101 process"),
            Err(err) => println!("Failed to open Wizard101 process: {}", err),
        });

        sleep(Duration::from_secs(2));

        let mut handles = HashMap::new();
        while handles.len() != target + initial_handles_len {
            handles = Self::get_all_wizard_handles();
            sleep(Duration::from_millis(500));
        }

        let new_handles: HashMap<_, _> = handles
            .into_iter()
            .filter(|(handle, _)| !initial_handles.contains_key(handle))
            .collect();

        for (i, handle) in new_handles.keys().enumerate() {
            let account = &self.config.accounts_data[i];
            self.wizard_login(*handle, &account.username, &account.password);
            Self::move_window(*handle, account.x_pos, account.y_pos);
        }
    }
}

fn main() {
    let config_path = "config.json";

    let config: Config = match std::fs::read_to_string(config_path) {
        Ok(content) => serde_json::from_str(&content).unwrap(),
        Err(err) => {
            println!("Error reading config file: {}", err);
            return;
        }
    };

    let app = Application::new(config);
    app.run();
}
