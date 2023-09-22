package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Select an action:")
		fmt.Println("1. Create a topic")
		fmt.Println("2. Subscribe to a topic")
		fmt.Println("3. Push a message to a topic")
		fmt.Println("0. Exit")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			createTopic(reader)
		case "2":
			subscribeToTopic(reader)
		case "3":
			pushMessageToTopic(reader)
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid input. Please try again.")
		}
	}
}

func createTopic(reader *bufio.Reader) {
	fmt.Print("Enter the topic name: ")
	topicName, _ := reader.ReadString('\n')
	topicName = strings.TrimSpace(topicName)

	data := url.Values{}
	data.Set("TopicName", topicName)

	resp, err := http.PostForm("http://localhost:8000/topic", data)
	if err != nil {
		fmt.Println("Error creating topic:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Topic created successfully")
	} else {
		fmt.Println("Error creating topic:", resp.Status)
	}
}

func subscribeToTopic(reader *bufio.Reader) {

	// 获取订阅主题
	fmt.Print("Enter the topic name: ")
	topicName, _ := reader.ReadString('\n')
	topicName = strings.TrimSpace(topicName)

	// 构建查询参数
	queryParams := url.Values{}
	queryParams.Set("TopicName", topicName)

	// 创建 SSE 连接
	url := "http://localhost:8000/subto?" + queryParams.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Failed to subscribe to topic:", resp.Status)
	}

	go readEvents(resp.Body)
	select {}
}

func readEvents(body io.Reader) {
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading event:", err)
		}

		fmt.Println("Received event:", line)
	}
}

func pushMessageToTopic(reader *bufio.Reader) {
	fmt.Print("Enter the topic name: ")
	topicName, _ := reader.ReadString('\n')
	topicName = strings.TrimSpace(topicName)

	fmt.Print("Enter the message to push: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	data := url.Values{}
	data.Set("TopicName", topicName)
	data.Set("Message", message)

	resp, err := http.PostForm("http://localhost:8000/pushto", data)
	if err != nil {
		fmt.Println("Error pushing message to topic:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Message pushed to topic successfully")
	} else {
		fmt.Println("Error pushing message to topic:", resp.Status)
	}
}
