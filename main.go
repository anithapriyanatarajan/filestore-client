package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const serverURL = "http://localhost:8080"

func main() {

	allowedOperations := []string{"add", "ls", "rm", "update", "wc", "freq-words"}

	if notInList(os.Args[1], allowedOperations) {
		fmt.Printf("%s is not a supported operation. Allowed Operations are %v\n", os.Args[1], allowedOperations)
		return
	}

	args := os.Args[2:]
	// 1.Add files + Incorporating hash logic for optimization
	if os.Args[1] == "add" {
		for i, arg := range args {
			fmt.Printf("%d: %s\n", i+1, arg)
			uploadFilePath := arg
			hash, err := generateFileHash(uploadFilePath)
			if err != nil {
				fmt.Println("Error generating file hash:", err)
				return
			}
			duplicationstatus, err := duplicateOnHashMatch(uploadFilePath, hash)
			if err != nil || duplicationstatus == "unmatched" {
				fmt.Println("WARN: performing hashmatch. proceeding standard upload ", err)
				err := uploadFile(uploadFilePath)
				if err != nil {
					fmt.Println("Error uploading file:", err)
					return
				}
			}
			fmt.Println("File uploaded successfully.")
		}
	}

	// 2. list files in store
	if os.Args[1] == "ls" {
		listFiles()
	}

	// 3. Remove a file
	if os.Args[1] == "rm" {
		deleteFile(os.Args[2])
	}

	// 4. update a file
	if os.Args[1] == "update" {
		updateFile(os.Args[2])
	}
}

func notInList(element string, list []string) bool {
	for _, value := range list {
		if value == element {
			return false // Element is present in the list
		}
	}
	return true // Element is not present in the list
}

func uploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/upload", serverURL)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Server responded with %s", response.Status)
	}

	return nil
}

func updateFile(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	url := fmt.Sprintf("%s/update/%s", serverURL, fileName)
	request, err := http.NewRequest("PUT", url, file)

	if err != nil {
		fmt.Println("Error creating PUT request:", err)
		return
	}

	// Set the content type header
	request.Header.Set("Content-Type", "application/octet-stream")

	// Create an HTTP client
	client := &http.Client{}

	// Perform the PUT request
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error performing PUT request:", err)
		return
	}
	defer response.Body.Close()

	// Print the response status code and body
	fmt.Println("Response Status Code:", response.Status)
	fmt.Println("Response Body:")
	io.Copy(os.Stdout, response.Body)
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

func generateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)
	return hashString, nil
}

func duplicateOnHashMatch(filename string, hash string) (string, error) {
	url := fmt.Sprintf("%s/verifyhashmatch/%s", serverURL, filename)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error listing files:", err)
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return "", err
		}
		fmt.Println("List of Files:")
		fmt.Println(string(body))
		return string(body), nil //http response should return unmatched
	} else {
		fmt.Println("List files failed. Server response:", response.Status)
		return "", err
	}
}
