package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

			matchingfile, err := findHashMatch(uploadFilePath, hash)
			if err != nil || matchingfile == "unmatched" {
				uploadFile(uploadFilePath, hash)
			} else {
				if !(uploadFilePath == matchingfile) {
					_, err := duplicatefile(uploadFilePath, matchingfile)
					if err != nil {
						fmt.Println("Error uploading file:", err)
						return
					}
				}
			}
			return
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

func uploadFile(filePath string, hashstring string) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return
	}

	// Add the hash as a URL value
	writer.WriteField("hash", hashstring)

	err = writer.Close()
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/upload", serverURL)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		fmt.Printf("File uploaded %s.\n", filePath)
		fmt.Println(string(body))
	} else {
		fmt.Println("could not upload file.", response.Status)
	}
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

//Calculate hash for the given file.

func generateFileHash(filePath string) (string, error) {
	fmt.Println("IsHashGenerated.")
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

func findHashMatch(filename string, hash string) (string, error) {
	serverURL := fmt.Sprintf("%s/findMatchingFile", serverURL)
	apiURL, err := url.Parse(serverURL)
	if err != nil {
		fmt.Println("Error parsing API URL:", err)
		return "", err
	}

	// Add query parameter for the hash
	parameters := url.Values{}
	parameters.Add("hash", hash)
	apiURL.RawQuery = parameters.Encode()

	// Send GET request
	response, err := http.Get(apiURL.String())
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return "", err
	}
	defer response.Body.Close()

	// Parse the JSON response
	var result map[string]string
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return "", err
	}

	// Print the result
	matchingFileName, exists := result["matchingFileName"]
	if !exists {
		return matchingFileName, nil
	} else {
		return "unmatched", nil
	}
}

func duplicatefile(uploadFilePath string, matchingfile string) (string, error) {
	serverURL := fmt.Sprintf("%s/copyFile", serverURL)
	apiURL, err := url.Parse(serverURL)
	if err != nil {
		fmt.Println("Error parsing API URL:", err)
		return "", err
	}

	// Add query parameter for the hash
	parameters := url.Values{}
	parameters.Add("src", matchingfile)
	parameters.Add("dest", uploadFilePath)
	apiURL.RawQuery = parameters.Encode()

	// Send GET request
	response, err := http.Get(apiURL.String())
	if err != nil {
		fmt.Println("Error making Copy request:", err)
		return "", err
	}
	defer response.Body.Close()

	// Parse the JSON response
	/*
		var result map[string]string
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			fmt.Println("Error decoding JSON response:", err)
			return "", err
		}
	*/

	return "", nil
}
