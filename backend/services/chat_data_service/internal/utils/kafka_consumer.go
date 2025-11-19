package utils

import (
	"encoding/json"
	"log"
	"time"

	"chat_data_service/internal/models"
	"chat_data_service/internal/services"

	"github.com/IBM/sarama"
)

func StartKafkaConsumer(brokers []string, topic string, service *services.ChatDataService) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Kafka consumer oluşturulamadı: %v", err)
	}
	defer consumer.Close()

	part, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Partition alınamadı: %v", err)
	}
	defer part.Close()

	log.Printf("Kafka dinleniyor: %s", topic)

	for msg := range part.Messages() {
		var data models.ChatMessage
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			log.Printf("Mesaj parse hatası: %v", err)
			continue
		}

		data.Timestamp = time.Now()
		if err := service.SaveMessage(&data); err != nil {
			log.Printf("Mesaj kaydedilemedi: %v", err)
		}
	}
}
