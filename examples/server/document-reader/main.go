package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nguyenthenguyen/docx"
)

type ReadFileArgs struct {
	FilePath string `json:"file_path" jsonschema:"Path to the file to read (supports .pdf, .docx, .doc, .txt, .md)"`
}

type ReadFileOutput struct {
	Content string `json:"content"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
}

func readPDF(filePath string) (string, error) {
	file, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	var content strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		texts := page.Content().Text
		for _, text := range texts {
			content.WriteString(text.S)
		}
		if pageNum < totalPages {
			content.WriteString("\n\n--- Page " + fmt.Sprintf("%d", pageNum) + " ---\n\n")
		}
	}

	return content.String(), nil
}

func readDocx(filePath string) (string, error) {
	doc, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open Word document: %w", err)
	}
	defer doc.Close()

	content := doc.Editable().GetContent()
	return content, nil
}

func readFile(ctx context.Context, req *mcp.CallToolRequest, args ReadFileArgs) (*mcp.CallToolResult, ReadFileOutput, error) {
	if args.FilePath == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "File path is required"},
			},
		}, ReadFileOutput{}, fmt.Errorf("file path required")
	}

	filePath := args.FilePath
	if !filepath.IsAbs(filePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to get current directory: %v", err)},
				},
			}, ReadFileOutput{}, err
		}
		filePath = filepath.Join(cwd, filePath)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("File not found: %s", args.FilePath)},
				},
			}, ReadFileOutput{}, fmt.Errorf("file not found: %s", args.FilePath)
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to access file: %v", err)},
			},
		}, ReadFileOutput{}, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var content string
	var fileType string

	switch ext {
	case ".txt", ".md":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to read file: %v", err)},
				},
			}, ReadFileOutput{}, err
		}
		content = string(data)
		fileType = "text"

	case ".pdf":
		pdfContent, err := readPDF(filePath)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to read PDF: %v", err)},
				},
			}, ReadFileOutput{}, err
		}
		content = pdfContent
		fileType = "pdf"

	case ".docx":
		docxContent, err := readDocx(filePath)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to read Word document: %v", err)},
				},
			}, ReadFileOutput{}, err
		}
		content = docxContent
		fileType = "docx"

	case ".doc":
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Old .doc format is not supported. Please convert to .docx format first"},
			},
		}, ReadFileOutput{}, fmt.Errorf("old .doc format not supported")

	default:
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Unsupported file type: %s. Supported types: .txt, .md, .pdf, .docx, .doc", ext)},
			},
		}, ReadFileOutput{}, fmt.Errorf("unsupported file type: %s", ext)
	}

	output := ReadFileOutput{
		Content: content,
		Size:    fileInfo.Size(),
		Type:    fileType,
	}

	resultText := fmt.Sprintf("File: %s\nSize: %d bytes\nType: %s\n\nContent:\n%s",
		args.FilePath, fileInfo.Size(), fileType, content)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultText},
		},
	}, output, nil
}

func main() {
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "document-reader",
		Version: "v1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read content from local files. Supports .txt, .md, .pdf, and .docx files",
	}, readFile)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
