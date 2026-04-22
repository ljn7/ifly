package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/ljn7/ifly/cli/internal/detect"
	"github.com/ljn7/ifly/cli/internal/updater"
	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

var updateDryRun bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update the ifly binary from the latest GitHub release",
	RunE: func(cmd *cobra.Command, args []string) error {
		httpc := &http.Client{Timeout: 30 * time.Second}
		client := updater.NewClient("https://api.github.com", repoOwner, repoName, httpc)
		rel, err := client.LatestRelease()
		if err != nil {
			return fmt.Errorf("fetch latest release: %w", err)
		}
		if !updater.IsNewer(cliVersion, rel.Tag) {
			cmd.Printf("already at %s (latest: %s)\n", cliVersion, rel.Tag)
			return nil
		}
		host := detect.Host()
		assetName := host.ReleaseAsset()
		asset := rel.AssetFor(assetName)
		if asset == nil {
			return fmt.Errorf("no asset for %s in release %s", assetName, rel.Tag)
		}
		sidecar := rel.AssetFor(assetName + ".sha256")
		if sidecar == nil {
			return fmt.Errorf("no .sha256 sidecar for %s", assetName)
		}

		cmd.Printf("downloading %s -> %s\n", rel.Tag, assetName)
		binary, err := updater.Download(asset.URL, httpc)
		if err != nil {
			return err
		}
		sumBytes, err := updater.Download(sidecar.URL, httpc)
		if err != nil {
			return err
		}
		if err := updater.VerifyChecksum(binary, string(sumBytes)); err != nil {
			return err
		}
		cmd.Println("checksum verified")

		if updateDryRun {
			cmd.Println("dry-run: not applying")
			return nil
		}
		if err := selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{}); err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		cmd.Printf("updated ifly to %s\n", rel.Tag)
		cmd.Println("plugin files NOT refreshed. run `ifly install --overwrite` to sync.")
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "download and verify but skip the replace")
	rootCmd.AddCommand(updateCmd)
}
