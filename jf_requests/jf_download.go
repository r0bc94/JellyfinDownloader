package jf_requests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func CreatePBar(length int64, description string) *progressbar.ProgressBar {
	desc := ""
	return progressbar.NewOptions64(
		length,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(1*time.Second),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionUseANSICodes(true),
	)
}

func DownloadFromUrl(downloadLink string, name string, outfile string, max int, current int) error {
	// --- 1. Set up the 'downloads' directory (Run once) ---
	const outputDir = "downloads"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %s", err)
	}
	outfile = filepath.Join(outputDir, outfile)

	// Initialize attempt counter
	attempt := 1

	// --- 2. RETRY LOOP ---
	for {
		// A. Check current file size
		var startByte int64 = 0
		if info, err := os.Stat(outfile); err == nil {
			startByte = info.Size()
		}

		// B. Make the Request
		req, _ := http.NewRequest("GET", downloadLink, nil)
		if startByte > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("\n[Attempt %d] Connection failed: %s. Retrying in 15s...\n", attempt, err)
			attempt++
			time.Sleep(15 * time.Second)
			continue
		}

		// C. Handle Response Codes
		if resp.StatusCode == 416 {
			resp.Body.Close()
			fmt.Printf("Skipping '%s' (Already complete)\n", outfile)
			return nil
		}

		if resp.StatusCode != 200 && resp.StatusCode != 206 {
			resp.Body.Close()
			fmt.Printf("\n[Attempt %d] Server error (%d). Retrying in 15s...\n", attempt, resp.StatusCode)
			attempt++
			time.Sleep(15 * time.Second)
			continue
		}

		// D. Open File
		flags := os.O_CREATE | os.O_WRONLY
		if resp.StatusCode == 206 {
			flags |= os.O_APPEND
			// Show attempt number on resume
			fmt.Printf("[Attempt %d] Resuming '%s' from %.2f MB\n", attempt, outfile, float64(startByte)/1024/1024)
		} else {
			startByte = 0
		}

		f, err := os.OpenFile(outfile, flags, 0644)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to open file: %s", err)
		}

		// E. Start Downloading
		bar := CreatePBar(resp.ContentLength, fmt.Sprintf("downloading %d/%d", current, max))
		_, copyErr := io.Copy(io.MultiWriter(f, bar), resp.Body)
		
		f.Close()
		resp.Body.Close()

		// F. Check for Success
		if copyErr == nil {
			return nil // Success!
		}

		// If we are here, the stream was interrupted
		fmt.Printf("\n[Attempt %d] Download interrupted: %s. Retrying in 15s...\n", attempt, copyErr)
		attempt++
		time.Sleep(15 * time.Second)
	}
}

func GetDownloadLinkForId(baseUrl string, token string, id string) string {
	return fmt.Sprintf(baseUrl+"/Items/%s/Download?api_key=%s", id, token)
}

func GetSuffixFromFilename(filename string) string {
	splittedFilename := strings.Split(filename, ".")
	suffix := splittedFilename[len(splittedFilename)-1]

	return suffix
}
