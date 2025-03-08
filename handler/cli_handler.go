package handler

import (
	"fmt"
	"os"
	"time"

	"github.com/mstgnz/self-hosted-url-shortener/service"
	"github.com/spf13/cobra"
)

// CLIHandler handles CLI commands
type CLIHandler struct {
	urlService *service.URLService
	baseURL    string
}

// NewCLIHandler creates a new CLI handler
func NewCLIHandler(urlService *service.URLService, baseURL string) *CLIHandler {
	return &CLIHandler{
		urlService: urlService,
		baseURL:    baseURL,
	}
}

// SetupCommands sets up the CLI commands
func (h *CLIHandler) SetupCommands() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "url-shortener",
		Short: "A self-hosted URL shortener",
		Long:  "A self-hosted URL shortener with SQLite backend",
	}

	// Shorten command
	shortenCmd := &cobra.Command{
		Use:   "shorten [url]",
		Short: "Shorten a URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			customCode, _ := cmd.Flags().GetString("code")
			h.shortenURL(args[0], customCode)
		},
	}
	shortenCmd.Flags().StringP("code", "c", "", "Custom short code")
	rootCmd.AddCommand(shortenCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all shortened URLs",
		Run: func(cmd *cobra.Command, args []string) {
			h.listURLs()
		},
	}
	rootCmd.AddCommand(listCmd)

	// Get command
	getCmd := &cobra.Command{
		Use:   "get [code]",
		Short: "Get details of a shortened URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			h.getURL(args[0])
		},
	}
	rootCmd.AddCommand(getCmd)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete [code]",
		Short: "Delete a shortened URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			h.deleteURL(args[0])
		},
	}
	rootCmd.AddCommand(deleteCmd)

	// QR command
	qrCmd := &cobra.Command{
		Use:   "qr [code]",
		Short: "Generate a QR code for a shortened URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			outputFile, _ := cmd.Flags().GetString("output")
			h.generateQR(args[0], outputFile)
		},
	}
	qrCmd.Flags().StringP("output", "o", "qr.png", "Output file for QR code")
	rootCmd.AddCommand(qrCmd)

	return rootCmd
}

// shortenURL shortens a URL
func (h *CLIHandler) shortenURL(longURL, customCode string) {
	url, err := h.urlService.ShortenURL(longURL, customCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)
	fmt.Printf("Short URL: %s\n", shortURL)
}

// listURLs lists all shortened URLs
func (h *CLIHandler) listURLs() {
	urls, err := h.urlService.ListURLs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(urls) == 0 {
		fmt.Println("No URLs found")
		return
	}

	fmt.Println("Shortened URLs:")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("%-10s %-15s %-30s %s\n", "ID", "Short Code", "Created", "Clicks")
	fmt.Println("------------------------------------------------------------")
	for _, url := range urls {
		fmt.Printf("%-10d %-15s %-30s %d\n", url.ID, url.ShortCode, url.CreatedAt.Format(time.RFC3339), url.Clicks)
	}
	fmt.Println("------------------------------------------------------------")
}

// getURL gets details of a shortened URL
func (h *CLIHandler) getURL(shortCode string) {
	url, err := h.urlService.GetURL(shortCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if url == nil {
		fmt.Println("URL not found")
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode)
	fmt.Println("URL Details:")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("ID:         %d\n", url.ID)
	fmt.Printf("Short Code: %s\n", url.ShortCode)
	fmt.Printf("Short URL:  %s\n", shortURL)
	fmt.Printf("Long URL:   %s\n", url.LongURL)
	fmt.Printf("Created:    %s\n", url.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Clicks:     %d\n", url.Clicks)
	fmt.Println("------------------------------------------------------------")
}

// deleteURL deletes a shortened URL
func (h *CLIHandler) deleteURL(shortCode string) {
	err := h.urlService.DeleteURL(shortCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("URL with code '%s' deleted successfully\n", shortCode)
}

// generateQR generates a QR code for a shortened URL
func (h *CLIHandler) generateQR(shortCode, outputFile string) {
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortCode)
	qrCode, err := h.urlService.GenerateQRCode(shortURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputFile, qrCode, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing QR code to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("QR code saved to %s\n", outputFile)
}
