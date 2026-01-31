//! Origin detection module - modular design for detecting where captured text came from.
//!
//! This module uses a trait-based design to allow easy swapping of detection strategies.
//! The main entry point is `detect()` which uses a CompositeDetector to try multiple strategies.

use crate::db::Origin;
use std::process::Command;
use thiserror::Error;

#[derive(Error, Debug)]
#[allow(dead_code)]
pub enum OriginError {
    #[error("Detection failed: {0}")]
    DetectionFailed(String),
    #[error("Script execution failed: {0}")]
    ScriptError(String),
}

/// Trait for origin detection strategies. Implement this to add new detection methods.
pub trait OriginDetector: Send + Sync {
    /// Attempt to detect the origin of captured text.
    /// Returns None if this detector cannot determine the origin.
    fn detect(&self) -> Option<Origin>;

    /// Human-readable name for logging
    fn name(&self) -> &'static str;
}

/// Detects origin from Safari browser
pub struct SafariDetector;

impl OriginDetector for SafariDetector {
    fn detect(&self) -> Option<Origin> {
        // Only detect if Safari is the frontmost app
        let script = r#"
            tell application "System Events"
                set frontApp to name of first application process whose frontmost is true
                if frontApp is "Safari" then
                    tell application "Safari"
                        if (count of windows) > 0 then
                            set currentTab to current tab of front window
                            return (URL of currentTab) & "||" & (name of currentTab)
                        end if
                    end tell
                end if
            end tell
            return ""
        "#;

        run_applescript(script).and_then(|output| {
            if output.is_empty() {
                return None;
            }
            let parts: Vec<&str> = output.split("||").collect();
            if parts.len() >= 2 && !parts[0].is_empty() {
                Some(Origin {
                    origin_type: "url".to_string(),
                    url: Some(parts[0].to_string()),
                    title: Some(parts[1].to_string()),
                    ..Default::default()
                })
            } else {
                None
            }
        })
    }

    fn name(&self) -> &'static str {
        "Safari"
    }
}

/// Detects origin from Google Chrome browser
pub struct ChromeDetector;

impl OriginDetector for ChromeDetector {
    fn detect(&self) -> Option<Origin> {
        // Only detect if Chrome is the frontmost app
        let script = r#"
            tell application "System Events"
                set frontApp to name of first application process whose frontmost is true
                if frontApp is "Google Chrome" then
                    tell application "Google Chrome"
                        if (count of windows) > 0 then
                            set activeTab to active tab of front window
                            return (URL of activeTab) & "||" & (title of activeTab)
                        end if
                    end tell
                end if
            end tell
            return ""
        "#;

        run_applescript(script).and_then(|output| {
            if output.is_empty() {
                return None;
            }
            let parts: Vec<&str> = output.split("||").collect();
            if parts.len() >= 2 && !parts[0].is_empty() {
                Some(Origin {
                    origin_type: "url".to_string(),
                    url: Some(parts[0].to_string()),
                    title: Some(parts[1].to_string()),
                    ..Default::default()
                })
            } else {
                None
            }
        })
    }

    fn name(&self) -> &'static str {
        "Chrome"
    }
}

/// Detects origin from Firefox browser
pub struct FirefoxDetector;

impl OriginDetector for FirefoxDetector {
    fn detect(&self) -> Option<Origin> {
        // Only detect if Firefox is the frontmost app
        // Firefox doesn't have great AppleScript support
        let script = r#"
            tell application "System Events"
                set frontApp to name of first application process whose frontmost is true
                if frontApp is "Firefox" then
                    -- Firefox's AppleScript support is limited
                    -- This may not work on all versions
                    return ""
                end if
            end tell
            return ""
        "#;

        run_applescript(script).and_then(|output| {
            if output.is_empty() {
                return None;
            }
            Some(Origin {
                origin_type: "url".to_string(),
                title: Some(output),
                ..Default::default()
            })
        })
    }

    fn name(&self) -> &'static str {
        "Firefox"
    }
}

/// Detects origin from Arc browser
pub struct ArcDetector;

impl OriginDetector for ArcDetector {
    fn detect(&self) -> Option<Origin> {
        // Only detect if Arc is the frontmost app
        let script = r#"
            tell application "System Events"
                set frontApp to name of first application process whose frontmost is true
                if frontApp is "Arc" then
                    tell application "Arc"
                        if (count of windows) > 0 then
                            set activeTab to active tab of front window
                            return (URL of activeTab) & "||" & (title of activeTab)
                        end if
                    end tell
                end if
            end tell
            return ""
        "#;

        run_applescript(script).and_then(|output| {
            if output.is_empty() {
                return None;
            }
            let parts: Vec<&str> = output.split("||").collect();
            if parts.len() >= 2 && !parts[0].is_empty() {
                Some(Origin {
                    origin_type: "url".to_string(),
                    url: Some(parts[0].to_string()),
                    title: Some(parts[1].to_string()),
                    ..Default::default()
                })
            } else {
                None
            }
        })
    }

    fn name(&self) -> &'static str {
        "Arc"
    }
}

/// Fallback detector that returns an unknown origin
pub struct FallbackDetector;

impl OriginDetector for FallbackDetector {
    fn detect(&self) -> Option<Origin> {
        Some(Origin {
            origin_type: "unknown".to_string(),
            ..Default::default()
        })
    }

    fn name(&self) -> &'static str {
        "Fallback"
    }
}

/// Composite detector that tries multiple strategies in order
pub struct CompositeDetector {
    detectors: Vec<Box<dyn OriginDetector>>,
}

impl CompositeDetector {
    pub fn new() -> Self {
        Self {
            detectors: vec![
                Box::new(SafariDetector),
                Box::new(ChromeDetector),
                Box::new(ArcDetector),
                Box::new(FirefoxDetector),
                Box::new(FallbackDetector),
            ],
        }
    }

    pub fn detect(&self) -> Origin {
        for detector in &self.detectors {
            log::debug!("Trying {} detector", detector.name());
            if let Some(origin) = detector.detect() {
                log::info!("Origin detected by {}: {:?}", detector.name(), origin.origin_type);
                return origin;
            }
        }
        // Should never reach here due to FallbackDetector
        Origin {
            origin_type: "unknown".to_string(),
            ..Default::default()
        }
    }
}

/// Run an AppleScript and return the trimmed output
fn run_applescript(script: &str) -> Option<String> {
    let output = Command::new("osascript")
        .arg("-e")
        .arg(script)
        .output()
        .ok()?;

    if output.status.success() {
        let result = String::from_utf8_lossy(&output.stdout).trim().to_string();
        Some(result)
    } else {
        log::debug!(
            "AppleScript failed: {}",
            String::from_utf8_lossy(&output.stderr)
        );
        None
    }
}

/// Main entry point for origin detection
pub async fn detect() -> Result<Origin, OriginError> {
    let detector = CompositeDetector::new();
    Ok(detector.detect())
}
