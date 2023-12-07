package kafka

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/hudangwei/common/logger"
	"go.uber.org/zap"
)

var (
	ErrProduceTimeout = errors.New("push message timeout")
	ErrProducerClosed = errors.New("producer has closed")
)

type KafkaProducerConfig struct {
	Brokers                []string `toml:"brokers"`
	Version                string   `toml:"version"`
	ChannelBufferSize      int      `toml:"channel_buffer_size"`
	ProducerFlushFrequency int      `toml:"producer_flush_frequency"`
}

type KafkaProducer struct {
	config       *KafkaProducerConfig
	producer     sarama.AsyncProducer
	client       sarama.Client
	flushTimeout time.Duration
	closeOnce    sync.Once
	closeFlag    int32
	closeChan    chan struct{}
	wg           sync.WaitGroup
}

func newSaramaConfig(producerConfig *KafkaProducerConfig) *sarama.Config {
	version, err := sarama.ParseKafkaVersion(producerConfig.Version)
	if err != nil {
		panic(fmt.Sprintf("Error parsing Kafka version: %v", err))
	}
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.ChannelBufferSize = producerConfig.ChannelBufferSize
	config.Net.MaxOpenRequests = 2

	config.Producer.MaxMessageBytes = 32 * 1024 * 1024
	config.Producer.Flush.Bytes = 32 * 1024 * 1024
	config.Producer.Flush.Messages = 512
	config.Producer.Flush.Frequency = time.Duration(producerConfig.ProducerFlushFrequency) * time.Millisecond
	config.Producer.Flush.MaxMessages = 2048

	config.Metadata.RefreshFrequency = 5 * time.Minute
	// config.Version = sarama.V2_0_0_0
	config.Version = version
	config.ClientID = "security_kafka"

	return config
}

func NewKafkaProducer(config *KafkaProducerConfig) *KafkaProducer {
	ret := &KafkaProducer{
		config:    config,
		closeChan: make(chan struct{}),
	}

	saramaConf := newSaramaConfig(config)
	client, err := sarama.NewClient(config.Brokers, saramaConf)
	if err != nil {
		panic(err)
	}
	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		client.Close()
		panic(err)
	}
	ret.client = client
	ret.producer = producer
	ret.flushTimeout = time.Duration(config.ProducerFlushFrequency)*time.Millisecond + time.Millisecond

	ret.wg.Add(2)
	go ret.handleError()
	go ret.handleSuccess()

	return ret
}

func (kp *KafkaProducer) IsClosed() bool {
	return atomic.LoadInt32(&kp.closeFlag) == 1
}

func (kp *KafkaProducer) Stop() {
	kp.closeOnce.Do(func() {
		atomic.StoreInt32(&kp.closeFlag, 1)
		close(kp.closeChan)
		if kp.producer != nil {
			kp.producer.AsyncClose()
		}
		if kp.client != nil {
			kp.client.Close()
		}
		kp.wg.Wait()
	})
}

func (kp *KafkaProducer) handleSuccess() {
	defer kp.wg.Done()
	for {
		select {
		case <-kp.closeChan:
			return
		case producerMessage := <-kp.producer.Successes():
			if kp.IsClosed() {
				return
			}
			if producerMessage != nil {

			}
		}
	}
}

func (kp *KafkaProducer) handleError() {
	defer kp.wg.Done()
	for {
		select {
		case <-kp.closeChan:
			return
		case err := <-kp.producer.Errors():
			if kp.IsClosed() {
				return
			}
			topic := err.Msg.Topic
			partition := strconv.FormatInt(int64(err.Msg.Partition), 10)
			key, _ := err.Msg.Key.Encode()
			switch err.Err {
			case sarama.ErrOutOfBrokers, sarama.ErrNotConnected, sarama.ErrShuttingDown:
				// 日志监控此语句，作为报警
				logger.Warn("async/producer Kafka Broker Unreachable", zap.String("topic", topic), zap.Any("brokers", kp.config.Brokers))
			case sarama.ErrLeaderNotAvailable, sarama.ErrNotLeaderForPartition, sarama.ErrControllerNotAvailable:
				if err := kp.client.RefreshMetadata(topic); err != nil {
					logger.Error("async/producer refresh metadata failed when leader unavailable", zap.String("topic", topic), zap.Any("brokers", kp.config.Brokers), zap.Error(err))
				}
			default:
				logger.Error("async/producer produce error", zap.String("topic", topic), zap.String("partition", partition), zap.Any("key", key), zap.Error(err))
			}
		}
	}
}

func (kp *KafkaProducer) Send(topic string, key string, msg []byte) error {
	if kp.IsClosed() {
		return ErrProducerClosed
	}
	producerMsg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msg),
	}
	var err error
	select {
	case kp.producer.Input() <- producerMsg:
		return nil
	case <-kp.closeChan:
		return ErrProducerClosed
	case <-time.After(kp.flushTimeout):
		err = ErrProduceTimeout
		if err1 := kp.client.RefreshMetadata(topic); err1 != nil {
			logger.Warn("async/producer refresh kafka metadata with error", zap.String("topic", topic), zap.Error(err1))
		} else {
			// retry once
			select {
			case kp.producer.Input() <- producerMsg:
				err = nil
			default:
			}
		}
	}
	return err
}
