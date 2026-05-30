package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"strings"

	"github.com/spf13/cobra"
	"github.com/zerebos/hindsight/internal/app"
	"github.com/zerebos/hindsight/internal/ingestion"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "hindsight",
		Short: "Personal browser history intelligence",
	}

	root.AddCommand(
		discoverCmd(),
		syncCmd(),
		statsCmd(),
		searchCmd(),
		sourcesCmd(),
		infoCmd(),
	)

	return root
}

// newApp initializes the App and fatally exits on error.
// Used by every command as its first step.
func newApp() *app.App {
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	return a
}

// ----------------------------------------------------------------
// hindsight discover
// ----------------------------------------------------------------

func discoverCmd() *cobra.Command {
	var register bool

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Scan for installed browsers and profiles",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			configDir, dataDir := a.Dirs()
			fmt.Printf("config dir: %s\n", configDir)
			fmt.Printf("data dir:   %s\n\n", dataDir)

			if register {
				fmt.Println("discovering and registering sources...")
				sources, errs := a.DiscoverAndRegister(ctx)
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "warning: %v\n", e)
				}
				fmt.Printf("registered %d source(s)\n", len(sources))
				for _, s := range sources {
					fmt.Printf("  [%d] %s\n", s.ID, s.Label.String)
				}
				return
			}

			fmt.Println("scanning for browsers (dry run -- use --register to save)...")
			detected, errs := a.DiscoverSources()
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "warning: %v\n", e)
			}
			if len(detected) == 0 {
				fmt.Println("no browsers found")
				return
			}
			fmt.Printf("found %d profile(s):\n", len(detected))
			for _, d := range detected {
				fmt.Printf("  [%s] %s\n", d.Browser, d.Label)
				fmt.Printf("        profile: %s\n", d.Profile)
				fmt.Printf("        path:    %s\n", d.Path)
			}
		},
	}

	cmd.Flags().BoolVarP(&register, "register", "r", false, "register found sources in the database")
	return cmd
}

// ----------------------------------------------------------------
// hindsight sync
// ----------------------------------------------------------------

func syncCmd() *cobra.Command {
	var sourceID int64

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync history from registered browser sources",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			if sourceID != 0 {
				fmt.Printf("syncing source %d...\n", sourceID)
				result, err := a.SyncSource(ctx, sourceID)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				printSyncResults([]ingestion.SyncResult{result})
				return
			}

			// Auto-discover and register if no sources exist yet
			sources, err := a.GetSources(ctx)
			if err != nil {
				log.Fatalf("get sources: %v", err)
			}
			if len(sources) == 0 {
				fmt.Println("no sources registered -- running discovery first...")
				registered, errs := a.DiscoverAndRegister(ctx)
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "warning: %v\n", e)
				}
				fmt.Printf("registered %d source(s)\n\n", len(registered))
			}

			fmt.Println("syncing all sources...")
			results := a.SyncAll(ctx)
			printSyncResults(results)
		},
	}

	cmd.Flags().Int64VarP(&sourceID, "source", "s", 0, "sync a specific source by ID")
	return cmd
}

func printSyncResults(results []ingestion.SyncResult) {
	totalNew, totalSkipped := 0, 0
	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("  source %-4d  error: %v\n", r.SourceID, r.Error)
		} else {
			fmt.Printf("  source %-4d  new: %-8d skipped: %d\n",
				r.SourceID, r.NewVisits, r.Skipped)
		}
		totalNew += r.NewVisits
		totalSkipped += r.Skipped
	}
	fmt.Printf("\ntotal  new: %d  skipped: %d\n", totalNew, totalSkipped)
}

// ----------------------------------------------------------------
// hindsight stats
// ----------------------------------------------------------------

func statsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Print dashboard summary statistics",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			filter := app.DashboardFilter{}
			if days > 0 {
				filter.StartTime = time.Now().AddDate(0, 0, -days).UnixMilli()
			}

			rangeLabel := "all time"
			if days > 0 {
				rangeLabel = fmt.Sprintf("last %d days", days)
			}

			// Summary stats
			stats, err := a.GetDashboardStats(ctx, filter)
			if err != nil {
				log.Fatalf("stats: %v", err)
			}
			fmt.Printf("[summary - %s]\n", rangeLabel)
			fmt.Printf("  total visits:   %d\n", stats.TotalVisits)
			fmt.Printf("  unique domains: %d\n", stats.UniqueDomains)
			fmt.Printf("  active days:    %d\n\n", stats.ActiveDays)

			// Top 10 domains
			domains, err := a.GetTopDomains(ctx, filter, 10)
			if err != nil {
				log.Fatalf("top domains: %v", err)
			}
			fmt.Printf("[top domains]\n")
			for i, d := range domains {
				fmt.Printf("  %2d. %-40s %d\n", i+1, d.Host, d.TotalVisits)
			}

			// Browser breakdown
			breakdown, err := a.GetBrowserBreakdown(ctx, filter)
			if err != nil {
				log.Fatalf("browser breakdown: %v", err)
			}
			fmt.Printf("\n[browser breakdown]\n")
			for _, b := range breakdown {
				blabel := b.Label.String
				if blabel == "" {
					blabel = b.Browser
				}
				fmt.Printf("  %-40s %d\n", blabel, b.TotalVisits)
			}

			// Heatmap peak hours
			cells, err := a.GetActivityHeatmap(ctx, filter, time.Local)
			if err != nil {
				log.Fatalf("heatmap: %v", err)
			}
			dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
			type peakCell struct {
				hour  int
				total int64
			}
			peaks := make(map[int]peakCell)
			for _, c := range cells {
				if ex, ok := peaks[c.Day]; !ok || c.TotalVisits > ex.total {
					peaks[c.Day] = peakCell{c.Hour, c.TotalVisits}
				}
			}
			fmt.Printf("\n[peak browsing hours]\n")
			for d := 0; d <= 6; d++ {
				if p, ok := peaks[d]; ok {
					fmt.Printf("  %s  %02d:00  (%d visits)\n", dayNames[d], p.hour, p.total)
				}
			}
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 0, "limit to last N days (0 = all time)")
	return cmd
}

// ----------------------------------------------------------------
// hindsight search
// ----------------------------------------------------------------

func searchCmd() *cobra.Command {
	var domain string
	var days int
	var page int64
	var pageSize int64

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search visit history",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			text := ""
			if len(args) > 0 {
				text = args[0]
			}

			params := app.SearchParams{
				Text:     text,
				Domain:   domain,
				Page:     page,
				PageSize: pageSize,
			}
			if days > 0 {
				params.StartTime = time.Now().AddDate(0, 0, -days).UnixMilli()
			}

			results, err := a.SearchVisits(ctx, params)
			if err != nil {
				log.Fatalf("search: %v", err)
			}

			fmt.Printf("%d results (page %d, showing %d)\n\n",
				results.Total, results.Page+1, len(results.Visits))

			for _, v := range results.Visits {
				title := v.Title.String
				if title == "" {
					title = "(no title)"
				}
				ts := time.UnixMilli(v.VisitedAt).Local().Format("2006-01-02 15:04")
				fmt.Printf("  %s  %s\n", ts, v.Domain)
				fmt.Printf("           %s\n", title)
				fmt.Printf("           %s\n\n", v.Url)
			}
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "filter by domain")
	cmd.Flags().IntVar(&days, "days", 0, "limit to last N days")
	cmd.Flags().Int64VarP(&page, "page", "p", 0, "page number (0-indexed)")
	cmd.Flags().Int64Var(&pageSize, "size", 20, "results per page")
	return cmd
}

// ----------------------------------------------------------------
// hindsight sources
// ----------------------------------------------------------------

func sourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage registered browser sources",
	}

	cmd.AddCommand(sourcesListCmd(), sourcesRemoveCmd())
	return cmd
}

func sourcesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered sources",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			sources, err := a.GetSources(ctx)
			if err != nil {
				log.Fatalf("get sources: %v", err)
			}

			if len(sources) == 0 {
				fmt.Println("no sources registered")
				fmt.Println("run: hindsight discover --register")
				return
			}

			fmt.Printf("%-4s  %-10s  %-30s  %-20s  %-20s  %s\n",
				"ID", "Browser", "Label", "Last Synced", "Last Visit Seen", "Error")
			fmt.Println(strings.Repeat("-", 110))

			for _, s := range sources {
				lastSynced := "never"
				if s.LastSyncedAt.Valid {
					lastSynced = time.UnixMilli(s.LastSyncedAt.Int64).Local().Format("2006-01-02 15:04")
				}

				lastSeen := "never"
				if s.LastVisitSeen.Valid {
					lastSeen = time.UnixMilli(s.LastVisitSeen.Int64).Local().Format("2006-01-02 15:04")
				}

				errStr := ""
				if s.LastError.Valid {
					// Truncate long errors for table display
					e := s.LastError.String
					if len(e) > 30 {
						e = e[:27] + "..."
					}
					errStr = e
				}

				fmt.Printf("%-4d  %-10s  %-30s  %-20s  %-20s  %s\n",
					s.ID, s.Browser, s.Label.String, lastSynced, lastSeen, errStr)
			}
		},
	}
}

func sourcesRemoveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a registered source by ID",
		Long: `Remove a source from the database by its ID.

Existing visits from this source are preserved in the database but will
no longer be associated with an active source. Use 'sources list' to
find the source ID.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			var id int64
			if _, err := fmt.Sscan(args[0], &id); err != nil {
				log.Fatalf("invalid source ID %q: %v", args[0], err)
			}

			// Show what we're about to remove
			sources, err := a.GetSources(ctx)
			if err != nil {
				log.Fatalf("get sources: %v", err)
			}

			var target *struct{ label, path string }
			for _, s := range sources {
				if s.ID == id {
					target = &struct{ label, path string }{s.Label.String, s.Path}
					break
				}
			}

			if target == nil {
				log.Fatalf("source %d not found", id)
			}

			if !force {
				fmt.Printf("remove source %d: %s\n", id, target.label)
				fmt.Printf("path: %s\n", target.path)
				fmt.Print("confirm? [y/N] ")
				var confirm string
				fmt.Scan(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("cancelled")
					return
				}
			}

			if err := a.RemoveSource(ctx, id); err != nil {
				log.Fatalf("remove source: %v", err)
			}
			fmt.Printf("removed source %d (%s)\n", id, target.label)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

// ----------------------------------------------------------------
// hindsight info
// ----------------------------------------------------------------

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show database and configuration info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			a := newApp()
			defer a.Close()

			configDir, dataDir := a.Dirs()
			fmt.Printf("config dir:      %s\n", configDir)
			fmt.Printf("data dir:        %s\n\n", dataDir)

			info, err := a.GetDBInfo(ctx)
			if err != nil {
				log.Fatalf("db info: %v", err)
			}

			fmt.Printf("database\n")
			fmt.Printf("  path:          %s\n", info.Path)
			fmt.Printf("  size:          %s\n", formatBytes(info.SizeBytes))
			fmt.Printf("  schema:        %s\n", info.SchemaVersion)
			fmt.Printf("  visits:        %d\n", info.TotalVisits)
			fmt.Printf("  domains:       %d\n", info.TotalDomains)
			fmt.Printf("  sources:       %d\n\n", info.TotalSources)

			cfg := a.GetSettings()
			fmt.Printf("settings\n")
			fmt.Printf("  sync interval: %d minutes\n", cfg.Sync.IntervalMinutes)
			fmt.Printf("  sync on launch:%v\n", cfg.Sync.SyncOnLaunch)
			fmt.Printf("  sync on wake:  %v\n", cfg.Sync.SyncOnWake)
			fmt.Printf("  launch at login:%v\n", cfg.General.LaunchAtLogin)
			fmt.Printf("  theme:         %s\n", cfg.General.Theme)
		},
	}
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}