package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	odp "github.com/patent-dev/uspto-odp"
)

func extractFilePathFromURI(downloadURI, productID string) string {
	if downloadURI == "" || productID == "" {
		return ""
	}

	pattern := "/products/files/" + productID + "/"
	index := strings.Index(downloadURI, pattern)
	if index == -1 {
		return ""
	}

	start := index + len(pattern)
	if start >= len(downloadURI) {
		return ""
	}

	return downloadURI[start:]
}

func demoBulk(ctx context.Context, client *odp.Client, reader *bufio.Reader) {
	printHeader("Bulk Data Download")

	productID := selectProduct(ctx, client, reader)
	if productID == "" {
		return
	}

	fileName := selectFile(ctx, client, reader, productID)
	if fileName == "" {
		return
	}

	downloadSelectedFile(ctx, client, productID, fileName)
}

func selectProduct(ctx context.Context, client *odp.Client, reader *bufio.Reader) string {
	fmt.Println("\nFetching all bulk data products...")

	result, err := client.SearchBulkProducts(ctx, "", 0, 100)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	if result.Count == nil || *result.Count == 0 {
		fmt.Println("No products found.")
		return ""
	}

	fmt.Printf("\nFound %d bulk data products:\n", *result.Count)
	fmt.Println("----------------------------------------")

	products := make(map[int]string)
	if result.BulkDataProductBag != nil {
		for i, product := range *result.BulkDataProductBag {
			if product.ProductIdentifier != nil && product.ProductTitleText != nil {
				fmt.Printf("%2d. %-15s : %s", i+1, *product.ProductIdentifier, *product.ProductTitleText)
				if product.ProductFrequencyText != nil {
					fmt.Printf(" (%s)", *product.ProductFrequencyText)
				}
				fmt.Println()
				products[i+1] = *product.ProductIdentifier
			}
		}
	}

	fmt.Print("\nSelect product number (or 'q' to quit): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "q" || choice == "Q" {
		return ""
	}

	num, err := strconv.Atoi(choice)
	if err != nil || num < 1 || num > len(products) {
		fmt.Println("Invalid selection")
		return ""
	}

	return products[num]
}

func selectFile(ctx context.Context, client *odp.Client, reader *bufio.Reader, productID string) string {
	fmt.Printf("\nFetching files for %s...\n", productID)

	result, err := client.GetBulkProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	if result.BulkDataProductBag == nil || len(*result.BulkDataProductBag) == 0 {
		fmt.Println("No product data found.")
		return ""
	}

	product := (*result.BulkDataProductBag)[0]
	if product.ProductFileBag == nil || product.ProductFileBag.FileDataBag == nil {
		fmt.Println("No files found for this product.")
		return ""
	}

	files := *product.ProductFileBag.FileDataBag
	totalFiles := len(files)

	if totalFiles == 0 {
		fmt.Println("No files available.")
		return ""
	}

	fmt.Printf("\nTotal files available: %d\n", totalFiles)

	pageSize := 20
	currentPage := 0

	for {
		start := currentPage * pageSize
		end := min(start+pageSize, totalFiles)

		fmt.Printf("\n=== Showing files %d-%d of %d ===\n", start+1, end, totalFiles)
		fmt.Println("----------------------------------------")

		fileMap := make(map[int]string)
		for i := start; i < end; i++ {
			file := files[i]
			if file.FileName != nil {
				displayNum := i - start + 1
				fmt.Printf("%2d. %s", displayNum, *file.FileName)
				if file.FileSize != nil {
					sizeMB := *file.FileSize / 1024 / 1024
					fmt.Printf(" (%.2f MB)", sizeMB)
				}
				if file.FileReleaseDate != nil {
					if len(*file.FileReleaseDate) >= 10 {
						fmt.Printf(" - %s", (*file.FileReleaseDate)[:10])
					}
				}
				fmt.Println()

				filePath := *file.FileName
				if file.FileDownloadURI != nil {
					if extracted := extractFilePathFromURI(*file.FileDownloadURI, productID); extracted != "" {
						filePath = extracted
					}
				}
				fileMap[displayNum] = filePath
			}
		}

		fmt.Println("\nOptions:")
		fmt.Println("  1-20: Select file number")
		if currentPage > 0 {
			fmt.Println("  p: Previous page")
		}
		if end < totalFiles {
			fmt.Println("  n: Next page")
		}
		fmt.Println("  s: Search for file")
		fmt.Println("  q: Quit")

		fmt.Print("\nYour choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "q", "Q":
			return ""
		case "n", "N":
			if end < totalFiles {
				currentPage++
			}
		case "p", "P":
			if currentPage > 0 {
				currentPage--
			}
		case "s", "S":
			fmt.Print("Enter search term: ")
			search, _ := reader.ReadString('\n')
			search = strings.TrimSpace(strings.ToLower(search))

			fmt.Println("\nSearch results:")
			count := 0
			for i, file := range files {
				if file.FileName != nil && strings.Contains(strings.ToLower(*file.FileName), search) {
					count++
					fmt.Printf("%d. %s", i+1, *file.FileName)
					if file.FileSize != nil {
						sizeMB := *file.FileSize / 1024 / 1024
						fmt.Printf(" (%.2f MB)", sizeMB)
					}
					fmt.Println()
					if count >= 10 {
						fmt.Println("... (showing first 10 matches)")
						break
					}
				}
			}
			if count == 0 {
				fmt.Println("No files found matching your search.")
			} else {
				fmt.Print("\nEnter file number from search results (or press enter to continue): ")
				searchChoice, _ := reader.ReadString('\n')
				searchChoice = strings.TrimSpace(searchChoice)
				if searchChoice != "" {
					num, err := strconv.Atoi(searchChoice)
					if err == nil && num > 0 && num <= len(files) {
						file := files[num-1]
						if file.FileName != nil {
							filePath := *file.FileName
							if file.FileDownloadURI != nil {
								if extracted := extractFilePathFromURI(*file.FileDownloadURI, productID); extracted != "" {
									filePath = extracted
								}
							}
							return filePath
						}
					}
				}
			}
		default:
			num, err := strconv.Atoi(choice)
			if err == nil && fileMap[num] != "" {
				return fileMap[num]
			}
			fmt.Println("Invalid selection")
		}
	}
}

func downloadSelectedFile(ctx context.Context, client *odp.Client, productID, fileName string) {
	result, err := client.GetBulkProduct(ctx, productID)
	if err != nil {
		fmt.Printf("Error getting product details: %v\n", err)
		return
	}

	var fileDownloadURI string
	if result.BulkDataProductBag != nil && len(*result.BulkDataProductBag) > 0 {
		product := (*result.BulkDataProductBag)[0]
		if product.ProductFileBag != nil && product.ProductFileBag.FileDataBag != nil {
			for _, file := range *product.ProductFileBag.FileDataBag {
				if file.FileDownloadURI != nil {
					if extracted := extractFilePathFromURI(*file.FileDownloadURI, productID); extracted == fileName {
						fileDownloadURI = *file.FileDownloadURI
						break
					}
				}
			}
		}
	}

	if fileDownloadURI == "" {
		fmt.Printf("Could not find FileDownloadURI for %s\n", fileName)
		return
	}

	outputPath := filepath.Base(fileName)
	fmt.Printf("Saving to: %s\n", outputPath)
	fmt.Printf("Download URL: %s\n", fileDownloadURI)

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	startTime := time.Now()
	var lastProgress int64
	var lastUpdate time.Time

	err = client.DownloadBulkFileWithProgress(ctx, fileDownloadURI, file,
		func(bytesComplete, bytesTotal int64) {
			now := time.Now()
			if now.Sub(lastUpdate) > 500*time.Millisecond || bytesComplete-lastProgress > 5*1024*1024 {
				if bytesTotal > 0 {
					percent := float64(bytesComplete) * 100 / float64(bytesTotal)
					elapsed := now.Sub(startTime).Seconds()
					speed := float64(bytesComplete) / elapsed / 1024 / 1024
					remaining := float64(bytesTotal-bytesComplete) / (float64(bytesComplete) / elapsed)

					fmt.Printf("\rProgress: %.1f%% | %.2f/%.2f MB | Speed: %.2f MB/s | ETA: %.0fs     ",
						percent,
						float64(bytesComplete)/1024/1024,
						float64(bytesTotal)/1024/1024,
						speed,
						remaining)
				} else {
					fmt.Printf("\rDownloaded: %.2f MB     ", float64(bytesComplete)/1024/1024)
				}
				lastProgress = bytesComplete
				lastUpdate = now
			}
		})

	fmt.Println()

	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		os.Remove(outputPath)
		return
	}

	info, _ := file.Stat()
	elapsed := time.Since(startTime)
	avgSpeed := float64(info.Size()) / elapsed.Seconds() / 1024 / 1024

	fmt.Printf("Success: downloaded %s\n", outputPath)
	fmt.Printf("   Size: %.2f MB\n", float64(info.Size())/1024/1024)
	fmt.Printf("   Time: %.1f seconds\n", elapsed.Seconds())
	fmt.Printf("   Avg Speed: %.2f MB/s\n", avgSpeed)
}
