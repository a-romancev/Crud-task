package event

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

type KafkaConf struct {
	Servers []string `mapstructure:"servers"`
	Topic   string   `mapstructure:"topic"`
}

func NewProducer(conf KafkaConf) Producer {
	p := Producer{}
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	prd, err := sarama.NewSyncProducer(conf.Servers, config)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	p.producer = prd
	p.topic = conf.Topic
	return p
}

type Producer struct {
	topic    string
	producer sarama.SyncProducer
}

func (p Producer) Produce(m string) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(m),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}
