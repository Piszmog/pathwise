use std::process::Command;

#[cfg(not(debug_assertions))]
const VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(debug_assertions)]
const VERSION: &str = "dev";

fn main() {
    let status = Command::new("tailwindcss")
        .args([
            "-i",
            "./assets/styles/input.css",
            "-o",
            format!("./assets/css/output@{}.css", VERSION).as_str(),
        ])
        .status()
        .expect("Failed to compile tailwindcss");

    if !status.success() {
        panic!("Failed to compile tailwindcss");
    }
}
