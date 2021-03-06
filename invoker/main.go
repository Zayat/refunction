package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/burntsushi/toml"
	"github.com/ostenbom/refunction/invoker/messages"
	"github.com/ostenbom/refunction/invoker/storage"
	"github.com/ostenbom/refunction/invoker/types"
	"github.com/ostenbom/refunction/invoker/workerpool"
	log "github.com/sirupsen/logrus"
)

const defaultCouchDBAddress = "http://admin:specialsecretpassword@127.0.0.1:5984"
const defaultKafkaAddress = "172.17.0.1:9093"
const defaultActivationDBName = "whisk_local_activations"
const defaultFunctionDBName = "whisk_local_whisks"

var defaultPoolCofig = []workerpool.GroupConfig{workerpool.GroupConfig{
	Size:        4,
	Runtime:     "python",
	TargetLayer: "serverless-function.py",
}}

type Config struct {
	PoolConfig  []workerpool.GroupConfig `toml:"poolgroup"`
	CouchConfig CouchConfig              `toml:"couch"`
	KafkaConfig KafkaConfig              `toml:"kafka"`
}

type KafkaConfig struct {
	Address string
}

type CouchConfig struct {
	Address            string
	ActivationDatabase string `toml:"activation_db"`
	FunctionDatabase   string `toml:"function_db"`
}

func startInvoker() int {
	var (
		configFile    string
		invokerNumber int
	)

	flag.IntVar(&invokerNumber, "id", -1, "unique id for the invoker")
	flag.StringVar(&configFile, "config", "config.toml", "config toml")
	flag.Parse()

	var config Config
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		printError(fmt.Errorf("could not load config toml: %s", err))
		return 1
	}

	config.FillDefaults()

	if invokerNumber < 0 {
		printError(fmt.Errorf("Invoker must have a unique id assigned greater than 0"))
		return 1
	}
	invokerID := fmt.Sprintf("invoker%d", invokerNumber)
	log.Info(fmt.Sprintf("Invoker with id: %s starting", invokerID))

	messenger, err := messages.NewMessenger(invokerNumber, config.KafkaConfig.Address)
	if err != nil {
		printError(err)
		return 1
	}
	defer messenger.Close()

	log.Debug("Messenger initialized")

	functionStorage, err := storage.NewFunctionStorage(config.CouchConfig.Address, config.CouchConfig.FunctionDatabase, config.CouchConfig.ActivationDatabase)
	if err != nil {
		printError(fmt.Errorf("could not establish couch connection: %s", err))
		return 1
	}

	log.Debug("Function storage connected")

	healthStop := messenger.StartHealthPings(invokerNumber)
	defer func() {
		healthStop <- true
	}()

	// Start fixed group of workers.
	workers, err := workerpool.NewWorkerPool(config.PoolConfig)
	if err != nil {
		printError(err)
		return 1
	}
	defer workers.Close()

	log.Info("Invoker initialized")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	messageChan := make(chan *types.ActivationMessage)
	errorChan := make(chan error)

	go func() {
		for {
			message, err := messenger.GetActivation()
			if err != nil {
				errorChan <- fmt.Errorf("could not pull from invoker queue: %s", err)
				return
			}
			messageChan <- message
		}
	}()

	// Graceful stopping in infinite loop
	for {
		select {
		case <-stopChan:
			log.Info("Shutting Down")
			return 0
		case message := <-messageChan:
			go func() {
				err = consumeMessage(message, functionStorage, workers, messenger)
				if err != nil {
					log.Error(err)
				}
			}()
		case err := <-errorChan:
			printError(err)
			return 1
		}
	}

	// return 0
}

func consumeMessage(activation *types.ActivationMessage, functionStorage storage.FunctionStorage, workers *workerpool.WorkerPool, messenger *messages.Messenger) error {
	// Fetch required function
	function, err := functionStorage.GetFunction(activation.Action.Path, activation.Action.Name)
	if err != nil {
		return fmt.Errorf("could not get activation function: %s", err)
	}

	messageLogger := log.WithFields(log.Fields{
		"ID":   activation.ActivationID,
		"name": function.Name,
	})

	messageLogger.WithFields(log.Fields{
		"code": function.Executable.Code,
	}).Debug("fetched function")

	// Schedule function
	result, err := workers.Run(function, activation.Parameters)
	if err != nil {
		return fmt.Errorf("could not run function %s: %s", function.Name, err)
	}

	messageLogger.WithFields(log.Fields{
		"result": result,
	}).Debug("function run complete")

	// Send ack
	if activation.Blocking {
		go func() {
			err := messenger.SendResult(activation, function, result)
			if err != nil {
				log.Errorf("could not send result %s, %s: %s", function.Name, activation.ActivationID, err)
			}
			messageLogger.Debug("result sent")
		}()
	}

	go func() {
		err := messenger.SendCompletion(activation, function, result)
		if err != nil {
			log.Error(fmt.Errorf("could not send completion %s, %s: %s", function.Name, activation.ActivationID, err))
		}
		messageLogger.Debug("completion sent")
	}()

	go func() {
		err := functionStorage.StoreActivation(activation, function, result)
		if err != nil {
			log.Error(fmt.Errorf("could not store activation %s, %s: %s", function.Name, activation.ActivationID, err))
		}
		messageLogger.Debug("activation stored")
	}()

	return nil
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	exitCode := startInvoker()
	os.Exit(exitCode)
}

func printError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}

func (c *Config) FillDefaults() {
	if len(c.PoolConfig) == 0 {
		c.PoolConfig = defaultPoolCofig
	}
	if c.KafkaConfig.Address == "" {
		c.KafkaConfig.Address = defaultKafkaAddress
	}
	if c.CouchConfig.Address == "" {
		c.CouchConfig.Address = defaultCouchDBAddress
	}
	if c.CouchConfig.ActivationDatabase == "" {
		c.CouchConfig.ActivationDatabase = defaultActivationDBName
	}
	if c.CouchConfig.FunctionDatabase == "" {
		c.CouchConfig.FunctionDatabase = defaultFunctionDBName
	}
}
