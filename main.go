package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

const serverURL = "http://localhost:8080"

func main() {
	// File paths for testing
	uploadFilePath := "./example.txt"
	downloadFileName := "example.txt"
	updateFilePath := "./updated_example.txt"

	// File Upload
	uploadFile(uploadFilePath)

	// File Download
	downloadFile(downloadFileName)

	// File Update
	updateFile(downloadFileName, updateFilePath)

	// File Delete
	deleteFile(downloadFileName)

	// File List
	listFiles()
}

func uploadFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	url := fmt.Sprintf("%s/upload", serverURL)

	var body bytes.Buffer
	writer := io.MultiWriter(&body, file)

	request, err := http.NewRequest("POST", url, &body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	request.Header.Set("Content-Type", "multipart/form-data")
	request.Header.Set("Filename", filePath)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return
	}
	defer response.Body.Close()

	fmt.Println("Upload Response:", response.Status)
}

func downloadFile(fileName string) {
	url := fmt.Sprintf("%s/download/%s", serverURL, fileName)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return
		}

		fmt.Println("Download successful.")
	} else {
		fmt.Println("Download failed. Server response:", response.Status)
	}
}

func updateFile(fileName string, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	url := fmt.Sprintf("%s/update/%s", serverURL, fileName)

	response, err := http.Put(url, "application/octet-stream", file)
	if err != nil {
		fmt.Println("Error updating file:", err)
		return
	}
	defer response.Body.Close()

	fmt.Println("Update Response:", response.Status)
}

func deleteFile(fileName string) {
	url := fmt.Sprintf("%s/delete/%s", serverURL, fileName)

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}
	defer response.Body.Close()

	fmt.Println("Delete Response:", response.Status)
}

func listFiles() {
	url := fmt.Sprintf("%s/list", serverURL)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error listing files:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		fmt.Println("List of Files:")
		fmt.Println(string(body))
	} else {
		fmt.Println("List files failed. Server response:", response.Status)
	}
}
