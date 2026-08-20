// Package browser opens a generated file in the user's browser.
//
// It exists so that no other package has to care about window management: the
// analysis writes a file, and this decides how a human ends up looking at it.
package browser

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Open shows path in Chrome, reusing the tab that already has it.
//
// Reloading an existing tab rather than stacking a new one matters here because
// the same report is regenerated over and over while tuning parameters; without
// it a session ends with twenty tabs of the same file, all but one stale.
func Open(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	url := "file://" + abs

	if runtime.GOOS == "darwin" {
		if err := openChromeMac(url, filepath.Base(abs)); err == nil {
			return nil
		}
		// Chrome missing or scripting refused — any browser beats none.
		return exec.Command("open", url).Run()
	}
	return openElsewhere(url)
}

// openChromeMac focuses an existing tab for the file, or opens one.
//
// A tab counts as a match when its URL is the file URL or merely ends in the
// same file name. The looser test is deliberate: the same report is often
// already open through an IDE's built-in web server, and insisting on the exact
// file:// URL would leave the reader with two tabs of one document, only one of
// which just got refreshed.
func openChromeMac(url, name string) error {
	script := fmt.Sprintf(`
set target to %q
set fileName to %q
tell application "Google Chrome"
	set found to false
	repeat with w in windows
		set i to 0
		repeat with t in tabs of w
			set i to i + 1
			set u to URL of t
			if u is target or u contains ("/" & fileName) then
				tell t to reload
				set active tab index of w to i
				set index of w to 1
				set found to true
				exit repeat
			end if
		end repeat
		if found then exit repeat
	end repeat
	if not found then
		if (count of windows) is 0 then
			make new window
		end if
		tell window 1 to make new tab with properties {URL:target}
	end if
	activate
end tell`, url, name)

	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func openElsewhere(url string) error {
	for _, candidate := range [][]string{
		{"google-chrome", url}, {"chromium", url},
		{"xdg-open", url}, {"rundll32", "url.dll,FileProtocolHandler", url},
	} {
		if _, err := exec.LookPath(candidate[0]); err != nil {
			continue
		}
		return exec.Command(candidate[0], candidate[1:]...).Run()
	}
	return fmt.Errorf("no browser launcher found for %s", runtime.GOOS)
}
