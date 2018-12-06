#!/usr/bin/env bash
# See https://brevi.link/shell-style and https://explainshell.com
## Contains basic functions to determine which os the script is being executed in.
##
## Usage:
##  This is a library; it contains common functions, so won't do anything if executed as a script.
##  It is the same as the copy in the everything repo.
##

function isLinux() {
  [[ "$(uname -s)" == "Linux" ]]
}

function isMacOS() {
  [[ "$(uname -s)" == "Darwin" ]]
}

function isWindows() {
  ! (isLinux || isMacOS)
}

# Return the target platform used by worker package names built for this OS.
function getPlatformName() {
  if isLinux; then
    echo "linux"
  elif isMacOS; then
    echo "macos"
  elif isWindows; then
    echo "win32"
  else
    echo "ERROR: Unknown platform." >&2
    exit 1
  fi
}
