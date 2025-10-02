package logger

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type (
	logMessage struct {
		buf    *bytes.Buffer
		logger *zap.Logger
	}

	kafkaWriter struct {
		producer sarama.AsyncProducer
		topic    string
		console  *zap.Logger
		wg       sync.WaitGroup
		done     chan struct{}
		in       io.WriteCloser
		out      io.ReadCloser
	}
)

var (
	bufferPool = sync.Pool{
		New: func() any {
			b := bytes.NewBuffer(make([]byte, 0, 3000))
			return &logMessage{
				buf: b,
			}
		},
	}
)

func newLogMessage(logger *zap.Logger) *logMessage {
	m := bufferPool.Get().(*logMessage)
	m.logger = logger
	return m
}

func (l *logMessage) Encode() ([]byte, error) {
	return l.buf.Bytes(), nil
}

func (l *logMessage) Length() int {
	return l.buf.Len()
}

func (l *logMessage) String() string {
	return l.buf.String()
}

func (l *logMessage) Write(p []byte) (n int, err error) {
	return l.buf.Write(p)
}

func (l *logMessage) release() {
	defer l.logger.Debug("release")

	l.buf.Truncate(0)
	bufferPool.Put(l)
}

func kafkaVersion(versionStr string, console *zap.Logger) (version sarama.KafkaVersion) {

	if versionStr == "" {
		return sarama.V2_7_2_0
	}

	var err error

	if version, err = sarama.ParseKafkaVersion(versionStr); err != nil {

		msg := fmt.Sprintf(
			"cannot parse kafka version from %s will use default %s: err=%s",
			versionStr,
			sarama.V2_7_2_0,
			err.Error(),
		)
		console.Warn(msg, zap.Error(err))

		return sarama.V2_7_2_0
	}

	return

}

func newKafkaWriter(config *KafkaConfig, console *zap.Logger) (*kafkaWriter, error) {

	saramaConfig := sarama.NewConfig()

	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	saramaConfig.Version = kafkaVersion(config.KafkaVersion, console)

	err := saramaConfig.Validate()

	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducer(config.Brokers, saramaConfig)

	if err != nil {
		return nil, err
	}

	out, in := io.Pipe()

	writer := &kafkaWriter{
		done:     make(chan struct{}),
		producer: producer,
		topic:    config.Topic,
		console:  console,
		out:      out,
		in:       in,
	}

	writer.errors()
	writer.successes()
	writer.readLoop()

	return writer, nil
}

func (w *kafkaWriter) Write(p []byte) (n int, err error) {

	n, err = w.in.Write(p)

	if err != nil && errors.Is(err, io.ErrClosedPipe) {
		w.console.Debug("skip Write due to closed pipe")
		return len(p), nil
	}

	return

}

func (w *kafkaWriter) readLoop() {

	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		w.console.Debug("readLoop()+")
		defer w.console.Debug("readLoop()-")

		reader := bufio.NewReader(w.out)

		message := newLogMessage(w.console)

		for {

			l, isPrefix, err := reader.ReadLine()
			w.console.Debug("readLine()", zap.Bool("isPrefix", isPrefix), zap.Error(err))

			if err != nil && errors.Is(err, io.EOF) {
				w.console.Debug("kafka logger read loop finished with EOF")
				return
			}

			if err != nil && errors.Is(err, io.ErrClosedPipe) {
				w.console.Debug("kafka logger read loop finished with ClosedPipe")
				return
			}

			if err != nil {
				w.console.Error(
					fmt.Sprintf(
						"kafka logger read loop failed with unexpected error: err=%s",
						err.Error(),
					),
					zap.Error(err),
				)
				return
			}

			message.Write(l)

			if isPrefix {
				continue
			}

			select {

			case <-w.done:

				w.console.Debug(
					"kafka logger is closed",
					zap.Stringer("skippedMessage", message),
				)

				return

			case w.producer.Input() <- &sarama.ProducerMessage{
				Topic: w.topic,
				Value: message,
			}:
				w.console.Debug("send message")

				message = newLogMessage(w.console)

			}

		}

	}()

}

func (w *kafkaWriter) successes() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		for s := range w.producer.Successes() {
			message := s.Value.(*logMessage)
			message.release()
		}
	}()
}

// errors log undelivered messages to console
func (w *kafkaWriter) errors() {

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		w.console.Debug("errors()+")
		defer w.console.Debug("errors()-")

		for producerErr := range w.producer.Errors() {

			message := producerErr.Msg.Value.(*logMessage)

			w.console.Error(
				fmt.Sprintf("kafka log failed: err=%s", producerErr.Err.Error()),
				zap.Stringer("logEntry", message),
				zap.Error(producerErr.Err),
			)

			message.release()
		}

	}()

}

func (w *kafkaWriter) Close() error {
	close(w.done)

	_ = w.out.Close()

	err := w.producer.Close()

	w.console.Info("kafka logger producer closed", zap.Error(err))

	w.wg.Wait()

	w.console.Sync()

	return err
}
