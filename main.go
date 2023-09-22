package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Topic struct {
	Name        string
	Subscribers map[string]chan<- string
}

type PubSubServer struct {
	Topics map[string]*Topic
}

func NewPubSubServer() *PubSubServer {
	return &PubSubServer{
		Topics: make(map[string]*Topic),
	}
}

func (ps *PubSubServer) createTopicHandler(c *gin.Context) {
	topicName := c.PostForm("TopicName")
	if topicName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TopicName is required"})
		return
	}

	_, exists := ps.Topics[topicName]
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic already exists"})
		return
	}

	topic := &Topic{
		Name:        topicName,
		Subscribers: make(map[string]chan<- string),
	}

	ps.Topics[topicName] = topic

	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Topic created: %s", topicName)})
}

func (ps *PubSubServer) subscribeHandler(c *gin.Context) {
	topicName := c.Query("TopicName")
	if topicName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TopicName is required"})
		return
	}

	topic, exists := ps.Topics[topicName]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic does not exist"})
		return
	}

	messageChan := make(chan string)
	topic.Subscribers[c.ClientIP()] = messageChan

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	// 将响应写入到HTTP响应流
	flusher, _ := c.Writer.(http.Flusher)

	for {

		select {
		case message := <-messageChan:

			// 将消息写入响应流
			fmt.Fprintf(c.Writer, "data: %s\n", message)
			// 刷新响应流，将数据发送给客户端
			flusher.Flush()
		case <-c.Writer.CloseNotify():
			delete(topic.Subscribers, c.ClientIP())
			log.Printf("Subscriber disconnected: %s", c.ClientIP())
			return
		}
	}
}

func (ps *PubSubServer) pushToTopicHandler(c *gin.Context) {
	topicName := c.PostForm("TopicName")
	if topicName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TopicName is required"})
		return
	}

	topic, exists := ps.Topics[topicName]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic does not exist"})
		return
	}

	message := c.PostForm("Message")

	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	for _, subscriber := range topic.Subscribers {
		subscriber <- message
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message pushed to topic"})
}

func main() {
	pubSubServer := NewPubSubServer()
	router := gin.Default()

	router.POST("/topic", pubSubServer.createTopicHandler)
	router.GET("/subto", pubSubServer.subscribeHandler)
	router.POST("/pushto", pubSubServer.pushToTopicHandler)

	log.Println("PubSub server started on port 8000")
	log.Fatal(router.Run(":8000"))
}
