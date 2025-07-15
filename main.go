package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConvertPDFParams represents the parameters for PDF conversion
type ConvertPDFParams struct {
	PDFPath string `json:"pdfPath"`
	Density int    `json:"density,omitempty"`
	Quality int    `json:"quality,omitempty"`
	Page    int    `json:"page,omitempty"`    // 0 means all pages, positive number for specific page
}

// ConvertResult represents the conversion result
type ConvertResult struct {
	Pages []PageImage `json:"pages"`
}

// PageImage represents a single page image
type PageImage struct {
	PageNumber int    `json:"pageNumber"`
	Base64Data string `json:"base64Data"`
}

// checkImageMagick checks if ImageMagick is installed
func checkImageMagick() error {
	cmd := exec.Command("convert", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ImageMagick is not installed or not in PATH: %v", err)
	}
	return nil
}

// convertPDFToImages converts a PDF to PNG images and returns them as base64
func convertPDFToImages(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ConvertPDFParams]) (*mcp.CallToolResultFor[any], error) {
	// Check if ImageMagick is available
	if err := checkImageMagick(); err != nil {
		return nil, err
	}

	// Set default values
	density := params.Arguments.Density
	if density == 0 {
		density = 300
	}
	quality := params.Arguments.Quality
	if quality == 0 {
		quality = 100
	}

	// Check if PDF file exists
	if _, err := os.Stat(params.Arguments.PDFPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF file does not exist: %s", params.Arguments.PDFPath)
	}

	// Create temporary directory for output images
	tempDir, err := os.MkdirTemp("", "pdf-to-png-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Output path pattern for images
	outputPath := filepath.Join(tempDir, "page-%03d.png")

	// Build ImageMagick convert command
	args := []string{
		"-density", fmt.Sprintf("%d", density),
		"-quality", fmt.Sprintf("%d", quality),
	}

	// Add page selection if specified
	if params.Arguments.Page > 0 {
		// ImageMagick uses 0-based indexing, so subtract 1
		args = append(args, fmt.Sprintf("%s[%d]", params.Arguments.PDFPath, params.Arguments.Page-1))
	} else {
		args = append(args, params.Arguments.PDFPath)
	}

	args = append(args, outputPath)

	// Run ImageMagick convert command
	cmd := exec.Command("convert", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("convert failed: %v, output: %s", err, output)
	}

	// Read all generated PNG files
	files, err := filepath.Glob(filepath.Join(tempDir, "page-*.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to list output files: %v", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no images were generated from the PDF")
	}

	// Sort files to ensure correct order
	var pages []PageImage
	for i, file := range files {
		// Read the PNG file
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read image file %s: %v", file, err)
		}

		// Store raw PNG data
		pages = append(pages, PageImage{
			PageNumber: i + 1,
			Base64Data: base64.StdEncoding.EncodeToString(data),
		})
	}

	// Return as MCP content with base64 data
	contents := []mcp.Content{
		&mcp.TextContent{Text: fmt.Sprintf("Successfully converted %d page(s) from PDF to PNG", len(pages))},
	}

	// Add each page as separate image content
	for _, page := range pages {
		// Decode base64 to bytes
		imageData, err := base64.StdEncoding.DecodeString(page.Base64Data)
		if err != nil {
			// If decode fails, skip this image
			contents = append(contents, &mcp.TextContent{
				Text: fmt.Sprintf("Page %d: Failed to decode base64 data", page.PageNumber),
			})
			continue
		}

		contents = append(contents, &mcp.ImageContent{
			Data:     imageData,
			MIMEType: "image/png",
		})
	}

	return &mcp.CallToolResultFor[any]{
		Content: contents,
	}, nil
}

// GetPageCountParams represents the parameters for getting page count
type GetPageCountParams struct {
	PDFPath string `json:"pdfPath"`
}

// getPageCount gets the total number of pages in a PDF
func getPageCount(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetPageCountParams]) (*mcp.CallToolResultFor[any], error) {
	// Check if ImageMagick is available
	if err := checkImageMagick(); err != nil {
		return nil, err
	}

	// Check if PDF file exists
	if _, err := os.Stat(params.Arguments.PDFPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF file does not exist: %s", params.Arguments.PDFPath)
	}

	// Use identify command to get page count
	cmd := exec.Command("identify", "-format", "%n\\n", params.Arguments.PDFPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("identify failed: %v, output: %s", err, output)
	}

	// Parse the output to get page count
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no output from identify command")
	}

	// The first line should contain the page count
	pageCount, err := strconv.Atoi(lines[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse page count: %v", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Total pages: %d", pageCount)},
		},
	}, nil
}

func main() {
	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pdf-to-image-converter",
		Version: "1.0.0",
	}, nil)

	// Add PDF conversion tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "convert_pdf_to_images",
		Description: "Convert a PDF file to PNG images encoded as base64. Use 'page' parameter to convert specific page (1-based index)",
	}, convertPDFToImages)

	// Add page count tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_pdf_page_count",
		Description: "Get the total number of pages in a PDF file",
	}, getPageCount)

	// Run server
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
